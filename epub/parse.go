package epub

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/jphastings/recipes/epub/prompts"
	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/llm"
	"github.com/jphastings/recipes/utils"
	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/pirmd/epub"
)

func Parse(b formats.Bundle, o formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	if o.LLM == nil {
		return nil, nil, fmt.Errorf("the ePub parser requires an LLM to be configured")
	}
	if len(b) != 1 {
		return nil, nil, fmt.Errorf("ePub bundles must contain exactly one ePub filename")
	}
	filename := b[0]

	ext := path.Ext(filename)
	if ext != collectionExt {
		return nil, nil, fmt.Errorf("doesn't appear to be an ePub file")
	}

	e, err := epub.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't open ePub: %w", err)
	}

	cd := &formats.CollectionDetails{
		Filename: strings.TrimSuffix(filename, ext),
	}

	// Use the ePub's title, if it's in the standard place
	if i, err := e.Information(); err == nil {
		cd.Name = i.Title[0]
	}
	pe := make(chan formats.ParseEvent)
	go func(e *epub.Epub, pe chan<- formats.ParseEvent, llmC *llm.Connection) {
		defer e.Close()
		extractRecipes(e, pe, llmC)
		close(pe)
	}(e, pe, o.LLM)

	return pe, cd, err
}

func extractRecipes(e *epub.Epub, pe chan<- formats.ParseEvent, c *llm.Connection) {
	index, err := getIndexFiles(e)
	if err != nil {
		pe <- formats.ParseEvent{Err: err}
		return
	}

	pm, err := getPageRef(e, index)
	if err != nil {
		pe <- formats.ParseEvent{Err: err}
		return
	}

	var n int
	for _, fragMap := range pm {
		n += len(fragMap)
	}
	pe <- formats.ParseEvent{N: n}

	for filename, fragMap := range pm {
		f, err := e.OpenItem(filename)
		if err != nil {
			pe <- formats.ParseEvent{
				Err: fmt.Errorf("unable to open page within ePub (%s): %w", filename, err),
				I:   len(fragMap),
			}
			continue
		}

		doc, err := html.Parse(f)
		if err != nil {
			pe <- formats.ParseEvent{
				Err: fmt.Errorf("unable to read page within ePub (%s): %w", filename, err),
				I:   len(fragMap),
			}
			continue
		}

		fragments := makeRecipeHTMLFragments(doc, fragMap)

		for _, frag := range fragments {
			// TODO: Handle more than one recipe in a fragment?
			if recipe, err := extractRecipeFromFragment(c, frag.html, e.OpenItem); err == nil {
				pe <- formats.ParseEvent{Recipe: recipe, I: 1}
			} else if err != llm.ErrWasIgnored {
				pe <- formats.ParseEvent{Err: err, I: 1}
			}
		}
	}
}

func extractRecipeFromFragment(c *llm.Connection, html string, openFile func(string) (fs.File, error)) (formats.Recipe, error) {
	answers, err := c.MultiQuery(prompts.System, prompts.Ignore, html, prompts.Map)
	if err != nil {
		return nil, err
	}

	r := formats.InterchangeRecipe{}

	var errs error
	for ans := range answers {
		if ans.Err != nil {
			errs = errors.Join(errs, ans.Err)
			continue
		}

		// LLMs are surprisingly dumb at returning no response.
		if ans.Answer == "NULL" {
			ans.Answer = ""
		}

		switch ans.Key {
		case "title":
			r.Title = ans.Answer
		case "description":
			r.Description = ans.Answer
		case "notes":
			r.Notes = ans.Answer

		case "ingredients":
			r.Ingredients = titledLists(ans.Answer)
		case "instructions":
			r.Instructions = titledLists(ans.Answer)

		case "cooktime":
			if dur, err := formats.MaybeDuration(ans.Answer).Parse(); err != nil {
				r.CookTime = dur
			}
		case "preptime":
			if dur, err := formats.MaybeDuration(ans.Answer).Parse(); err != nil {
				r.PrepTime = dur
			}
		case "totaltime":
			if dur, err := formats.MaybeDuration(ans.Answer).Parse(); err != nil {
				r.TotalTime = dur
			}

		case "images":
			for _, fname := range lines(ans.Answer) {
				if file, err := openFile(fname); err != nil {
					r.Images = append(r.Images, file)
				}
			}

		case "yield":
			if yield, err := strconv.Atoi(ans.Answer); err != nil && yield != 0 {
				r.Yield = strconv.Itoa(yield)
			}

		default:
			panic(fmt.Sprintf("%s not handled", ans.Key))
		}
	}

	if vErrs := r.Validate(); len(vErrs) > 0 {
		return nil, errors.Join(fmt.Errorf("not a valid recipe"), errors.Join(vErrs...), errs)
	}

	return r, nil
}

func lines(in string) []string {
	all := strings.Split(in, "\n")
	for i, line := range all {
		all[i] = strings.TrimSpace(line)
	}
	return all
}

func titledLists(in string) []formats.TitledList {
	var lists []formats.TitledList

	list := formats.TitledList{}
	for _, line := range lines(in) {
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "#") {
			list.List = append(list.List, strings.Trim(line, "-*• :"))
			continue
		}

		if len(list.List) > 0 {
			lists = append(lists, list)
		}
		list = formats.TitledList{
			Title: strings.TrimSpace(line[1:]),
		}
	}

	if len(list.List) > 0 {
		lists = append(lists, list)
	}

	return lists
}

var otherIndexMatcher = regexp.MustCompile(`(?i)^(.*index.*)\d+(\.x?html)$`)

// Finds the "index" page of the ePub (ie. what a human would use to look up what recipe is on what page)
// Must guess if the recipe is spread across multiple HTML pages and, in that case, assumes that each of them has the same filename prefix and a numeric suffix
func getIndexFiles(e *epub.Epub) ([]string, error) {
	pkg, err := e.Package()
	if err != nil {
		return nil, fmt.Errorf("couldn't open ePub package file: %w", err)
	}

	var tocFile string
	for _, item := range pkg.Manifest.Items {
		if item.MediaType == "application/x-dtbncx+xml" {
			tocFile = item.Href
			break
		}
	}
	if tocFile == "" {
		return nil, fmt.Errorf("no table of contents listed in the ePub package")
	}

	f, err := e.OpenItem(tocFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open the table of contents file")
	}

	ncx, err := parseNCX(f)
	if err != nil {
		return nil, fmt.Errorf("couldn't read ePub's table of contents: %w", err)
	}

	var indexFiles []string
	var checkForOthers []string

	for _, p := range ncx.NavMap.NavPoints {
		text := strings.TrimSpace(strings.ToLower(p.Label.Text))
		if text == "index" {
			indexFiles = append(indexFiles, p.Content.Src)
			checkForOthers = otherIndexMatcher.FindStringSubmatch(p.Content.Src)
			break
		}
	}

	if checkForOthers == nil {
		return indexFiles, nil
	}

	otherMatcher, err := regexp.Compile(fmt.Sprintf(
		`^%s\d+%s$`,
		regexp.QuoteMeta(checkForOthers[1]),
		regexp.QuoteMeta(checkForOthers[2]),
	))
	if err != nil {
		return indexFiles, nil
	}

	for _, item := range pkg.Manifest.Items {
		if item.Href != checkForOthers[0] && otherMatcher.MatchString(item.Href) {
			indexFiles = append(indexFiles, item.Href)
		}
	}

	return indexFiles, nil
}

// index filename -> HTML id -> page number
type realPageMap map[string]map[string]utils.Pages

// Creates a map of the real page number of every reference to one in the provided index pages (likely a recipe), as well as the HTML fragment that it points to
func getPageRef(e *epub.Epub, indexFiles []string) (realPageMap, error) {
	pm := make(realPageMap)
	for _, idx := range indexFiles {
		f, err := e.OpenItem(idx)
		if err != nil {
			return nil, fmt.Errorf("unable to open Index (%s): %w", idx, err)
		}

		doc, err := htmlquery.Parse(f)
		if err != nil {
			return nil, fmt.Errorf("unable to parse HTML in index: %w", err)
		}

		links, err := htmlquery.QueryAll(doc, "//a[text()]")
		if err != nil {
			return nil, fmt.Errorf("unable to find page references in the index: %w", err)
		}

		for _, link := range links {
			destParts := strings.Split(getHrefAttribute(link), "#")
			// Ignore links that don't have a fragment
			if len(destParts) != 2 {
				continue
			}
			// Ignore links to the index
			destPath := path.Join(path.Dir(idx), destParts[0])
			if slices.Contains(indexFiles, destPath) {
				continue
			}
			text := htmlquery.InnerText(link)
			pages, err := utils.ParsePages(text)
			if err != nil {
				continue
			}

			if _, ok := pm[destPath]; !ok {
				pm[destPath] = make(map[string]utils.Pages)
			}
			pm[destPath][destParts[1]] = pages
		}
	}

	return pm, nil
}

// Pulls the href attribute value from an HTML node
func getHrefAttribute(node *html.Node) string {
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				return attr.Val
			}
		}
	}
	return ""
}

const fragmentTargetXPath = "//*[@id = '%s' or (local-name() = 'a' and @name = '%s')]"

type htmlFragmentWithPages struct {
	html  string
	pages utils.Pages
}

func makeRecipeHTMLFragments(doc *html.Node, fragMap map[string]utils.Pages) []htmlFragmentWithPages {
	var splitIndices []int
	idxMap := make(map[int]utils.Pages)
	docStr := htmlquery.OutputHTML(doc, true)

	for frag, pgs := range fragMap {
		node, err := htmlquery.Query(doc, fmt.Sprintf(fragmentTargetXPath, frag, frag))
		if err != nil {
			continue
		}

		idx := strings.Index(docStr, htmlquery.OutputHTML(node, true))
		idxMap[idx] = pgs

		splitIndices = append(splitIndices, idx)
	}

	sort.Ints(splitIndices)

	fragments := make([]htmlFragmentWithPages, len(splitIndices))
	for i, idx := range splitIndices {
		if i == len(splitIndices)-1 {
			fragments[i] = htmlFragmentWithPages{
				html:  docStr[idx:],
				pages: idxMap[idx],
			}
		} else {
			fragments[i] = htmlFragmentWithPages{
				html:  docStr[idx:splitIndices[i+1]],
				pages: idxMap[idx],
			}
		}
	}

	return fragments
}
