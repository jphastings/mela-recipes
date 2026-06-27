package schemaorg

import (
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"golang.org/x/net/html"
)

// microdataRecipe extracts a recipe expressed as schema.org microdata
// (itemscope / itemtype / itemprop attributes), returning any image URLs
// separately.
func microdataRecipe(doc *html.Node) (formats.InterchangeRecipe, []string, bool) {
	scope := findFirst(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			hasAttr(n, "itemscope") &&
			typeContainsRecipe(getAttr(n, "itemtype"))
	})
	if scope == nil {
		return formats.InterchangeRecipe{}, nil, false
	}
	ir, urls := buildFromProps(collectProps(scope))
	return ir, urls, true
}

func hasAttr(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			return true
		}
	}
	return false
}

func typeContainsRecipe(itemtype string) bool {
	for _, t := range strings.Fields(itemtype) {
		if strings.EqualFold(t, "Recipe") || strings.HasSuffix(strings.TrimRight(t, "/"), "/Recipe") {
			return true
		}
	}
	return false
}

// collectProps gathers the itemprop elements belonging to the recipe item,
// without descending into nested items (eg. an author or nutrition itemscope),
// whose own properties would otherwise be misattributed to the recipe.
func collectProps(scope *html.Node) map[string][]*html.Node {
	props := map[string][]*html.Node{}

	var walk func(n *html.Node, inNested bool)
	walk = func(n *html.Node, inNested bool) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != html.ElementNode {
				continue
			}
			if prop := getAttr(c, "itemprop"); prop != "" && !inNested {
				for _, p := range strings.Fields(prop) {
					props[p] = append(props[p], c)
				}
			}
			walk(c, inNested || hasAttr(c, "itemscope"))
		}
	}
	walk(scope, false)
	return props
}

func buildFromProps(props map[string][]*html.Node) (formats.InterchangeRecipe, []string) {
	ir := formats.NewInterchangeRecipe()

	ir.Title = firstPropText(props, "name")
	ir.Description = firstPropText(props, "description")
	ir.Yield = firstPropText(props, "recipeYield")

	for _, n := range propNodes(props, "recipeIngredient", "ingredients") {
		if s := propText(n); s != "" {
			ir.Ingredients = appendItem(ir.Ingredients, s)
		}
	}

	if steps := instructionSteps(props["recipeInstructions"]); len(steps) > 0 {
		ir.Instructions = []formats.TitledList{{List: steps}}
	}

	ir.PrepTime = parseISODuration(firstPropRaw(props, "prepTime"))
	ir.CookTime = parseISODuration(firstPropRaw(props, "cookTime"))
	ir.TotalTime = parseISODuration(firstPropRaw(props, "totalTime"))

	for _, key := range []string{"recipeCategory", "recipeCuisine", "keywords"} {
		for _, n := range props[key] {
			for _, part := range strings.Split(propText(n), ",") {
				if p := strings.TrimSpace(part); p != "" {
					ir.Tags = append(ir.Tags, p)
				}
			}
		}
	}
	ir.Tags = dedupe(ir.Tags)

	ir.Source = formats.Source{
		Name: firstPropText(props, "author"),
		URI:  firstPropRaw(props, "url"),
	}

	var urls []string
	for _, n := range props["image"] {
		if u := propRaw(n); u != "" {
			urls = append(urls, u)
		}
	}
	return ir, urls
}

// instructionSteps turns recipeInstructions microdata into a flat step list:
// several elements are taken as one step each; a single container is split on its
// block structure (eg. <li>/<p>).
func instructionSteps(nodes []*html.Node) []string {
	switch len(nodes) {
	case 0:
		return nil
	case 1:
		if lines := cleanLines(nodeText(nodes[0])); len(lines) > 0 {
			return lines
		}
		if s := propText(nodes[0]); s != "" {
			return []string{s}
		}
		return nil
	default:
		var steps []string
		for _, n := range nodes {
			if s := propText(n); s != "" {
				steps = append(steps, s)
			}
		}
		return steps
	}
}

// propText returns an itemprop's human-readable value.
func propText(n *html.Node) string {
	if v := getAttr(n, "content"); v != "" {
		return cleanText(v)
	}
	return cleanText(nodeText(n))
}

// propRaw returns an itemprop's machine value (for durations, URLs), preferring
// the content / datetime / href / src attributes over visible text.
func propRaw(n *html.Node) string {
	if v := getAttr(n, "content"); v != "" {
		return v
	}
	if v := getAttr(n, "datetime"); v != "" {
		return v
	}
	if n.Data == "a" || n.Data == "link" {
		if v := getAttr(n, "href"); v != "" {
			return v
		}
	}
	if n.Data == "img" {
		if v := getAttr(n, "src"); v != "" {
			return v
		}
	}
	return cleanText(nodeText(n))
}

func propNodes(props map[string][]*html.Node, keys ...string) []*html.Node {
	var out []*html.Node
	for _, k := range keys {
		out = append(out, props[k]...)
	}
	return out
}

func firstPropText(props map[string][]*html.Node, key string) string {
	if ns := props[key]; len(ns) > 0 {
		return propText(ns[0])
	}
	return ""
}

func firstPropRaw(props map[string][]*html.Node, key string) string {
	if ns := props[key]; len(ns) > 0 {
		return propRaw(ns[0])
	}
	return ""
}

func appendItem(lists []formats.TitledList, item string) []formats.TitledList {
	if len(lists) == 0 {
		return []formats.TitledList{{List: []string{item}}}
	}
	lists[0].List = append(lists[0].List, item)
	return lists
}
