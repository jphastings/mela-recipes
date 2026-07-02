package induce

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Section is a (possibly titled) group of ingredient or step lines, e.g.
// "For the sauce". An untitled section (Title == "") is the default group.
type Section struct {
	Title string   `json:"title,omitempty"`
	Items []string `json:"items"`
}

// Recipe is one captured recipe. Title/ingredient/step text is verbatim from the
// source DOM; interpretation only adds metadata (categories, yield) and strips
// recognised markers — it never rewrites the lexical core.
type Recipe struct {
	Title         string    `json:"title"`
	VerbatimTitle string    `json:"-"` // title before marker-stripping, for the gate
	Subtitle      string    `json:"subtitle,omitempty"`
	Description   string    `json:"description,omitempty"`
	Ingredients   []Section `json:"ingredients"`
	Steps         []Section `json:"steps"`
	Yield         string    `json:"yield,omitempty"`
	Categories    []string  `json:"categories,omitempty"`
	Pages         []string  `json:"pages,omitempty"`
	Images        []string  `json:"images,omitempty"` // candidate image item paths, nearest first
	Issues        []string  `json:"issues,omitempty"` // verification-gate failures
}

func (r Recipe) OK() bool             { return len(r.Issues) == 0 }
func (r Recipe) IngredientCount() int { return countItems(r.Ingredients) }
func (r Recipe) StepCount() int       { return countItems(r.Steps) }

func countItems(secs []Section) int {
	n := 0
	for _, s := range secs {
		n += len(s.Items)
	}
	return n
}

func addItem(secs *[]Section, item string) {
	if len(*secs) == 0 {
		*secs = append(*secs, Section{})
	}
	last := &(*secs)[len(*secs)-1]
	last.Items = append(last.Items, item)
}

func addToSection(secs *[]Section, t string, heading bool) {
	if heading {
		*secs = append(*secs, Section{Title: t})
	} else {
		addItem(secs, t)
	}
}

func yieldValue(t string) string {
	if n := firstNumber(t); n != "" {
		return n
	}
	return t
}

// mergeBareQuantity handles hanging-indent layouts where an ingredient's
// quantity and name sit in adjacent blocks ("1" | "cup (180g) tomatoes"). When
// block i is a bare quantity, it joins the following block's name onto it.
func mergeBareQuantity(blocks []*html.Node, i int, t string, titleNode, subtitleNode *html.Node, emphSignal bool) (string, bool) {
	if !isBareQuantity(t) || i+1 >= len(blocks) {
		return t, false
	}
	nb := blocks[i+1]
	if nb == titleNode || nb == subtitleNode {
		return t, false
	}
	nt := text(nb)
	if nt == "" || !hasLetters(nt) || isBareQuantity(nt) || runeLen(nt) >= 200 {
		return t, false
	}
	if emphSignal && emphasised(nb) && !hasNumeral(nt) {
		return t, false // the next block is a group heading, not a name
	}
	return t + " " + nt, true
}

func dropEmptySections(secs []Section) []Section {
	var out []Section
	for _, s := range secs {
		if len(s.Items) > 0 {
			out = append(out, s)
		}
	}
	return out
}

// Report is the outcome of applying a profile to a book.
type Report struct {
	Recipes    []Recipe
	Confidence Confidence
	Flagged    int
}

func (p *Profile) markerSels() []Sel {
	out := make([]Sel, len(p.Markers))
	for i, m := range p.Markers {
		out[i] = m.Sel
	}
	return out
}

var pageIDRe = regexp.MustCompile(`(?i)p(?:age)?[_-]?(\d+)`)

func pagesIn(u Unit) []string {
	seen := map[string]bool{}
	var out []string
	collect := func(root *html.Node) {
		eachElement(root, func(n *html.Node) {
			for _, a := range n.Attr {
				if a.Key != "id" {
					continue
				}
				if m := pageIDRe.FindStringSubmatch(a.Val); m != nil {
					if !seen[m[1]] {
						seen[m[1]] = true
						out = append(out, m[1])
					}
				}
			}
		})
	}
	for _, b := range u.Blocks {
		collect(b)
	}
	sort.Slice(out, func(i, j int) bool {
		a, _ := strconv.Atoi(out[i])
		b, _ := strconv.Atoi(out[j])
		return a < b
	})
	return out
}

func (p *Profile) buildRecipe(u Unit) Recipe {
	var r Recipe
	strip := p.markerSels()

	titleNode := firstMatch(u, p.Fields[RoleTitle])
	subtitleNode := firstMatch(u, p.Fields[RoleSubtitle])

	if titleNode != nil {
		r.VerbatimTitle = text(titleNode)
		r.Title = textExcluding(titleNode, strip)
	}
	if subtitleNode != nil {
		st := textExcluding(subtitleNode, strip)
		// A once-per-recipe line before the ingredients that carries a number
		// ("Makes 56", "Serves 4") is a yield, not a translated-name subtitle.
		if hasNumeral(st) && runeLen(st) <= 40 {
			r.Yield = yieldValue(st)
		} else {
			r.Subtitle = st
		}
	}

	// Interpretation: marker -> category (operating on already-captured nodes).
	cats := map[string]bool{}
	for _, node := range []*html.Node{titleNode, subtitleNode} {
		if node == nil {
			continue
		}
		for _, m := range p.Markers {
			eachElement(node, func(n *html.Node) {
				if m.Sel.Matches(n) && strings.EqualFold(text(n), m.Equals) {
					cats[m.Category] = true
				}
			})
		}
	}
	for c := range cats {
		r.Categories = append(r.Categories, c)
	}
	sort.Strings(r.Categories)

	if f, ok := p.Fields[RoleDescription]; ok {
		var ds []string
		for _, b := range u.Blocks {
			if f.Matches(b) {
				if t := text(b); t != "" {
					ds = append(ds, t)
				}
			}
		}
		r.Description = strings.Join(ds, "\n\n")
	}

	ingF, hasIng := p.Fields[RoleIngredients]
	stepF, hasStep := p.Fields[RoleSteps]

	// Emphasis separates headings/yield from items only when it's exceptional.
	// If a book emboldens every ingredient (e.g. Cooking for Geeks), emphasis
	// carries no signal here, so don't read yield/headings from it.
	emphSignal := true
	if hasIng {
		emph, total := 0, 0
		for _, b := range u.Blocks {
			if ingF.Matches(b) {
				total++
				if emphasised(b) {
					emph++
				}
			}
		}
		emphSignal = total == 0 || float64(emph)/float64(total) < 0.6
	}

	consumed := make([]bool, len(u.Blocks))
	active := &r.Ingredients
	for i, b := range u.Blocks {
		if consumed[i] || b == titleNode {
			continue
		}
		t := text(b)
		if t == "" {
			continue
		}
		// Yield: a wholly-emphasised line carrying a number ("Serves 4",
		// "Pour 4 personnes"). Emphasis + numeral is language-agnostic.
		if emphSignal && r.Yield == "" && emphasised(b) && hasNumeral(t) && runeLen(t) <= 40 {
			r.Yield = yieldValue(t)
			continue
		}
		if b == subtitleNode {
			continue
		}
		// A wholly-emphasised line without a number introduces a group
		// ("For the sauce"); quantity-bearing or plain lines stay items.
		heading := emphSignal && emphasised(b) && !hasNumeral(t) && runeLen(t) <= 80
		switch {
		case hasIng && ingF.Matches(b):
			active = &r.Ingredients
			if it, used := mergeBareQuantity(u.Blocks, i, t, titleNode, subtitleNode, emphSignal); used {
				t = it
				consumed[i+1] = true
			}
			addToSection(active, t, heading)
		case hasStep && stepF.Matches(b):
			active = &r.Steps
			addToSection(active, t, heading)
		case heading:
			// A distinct heading class (not itself a field) groups whichever
			// list we are currently filling.
			*active = append(*active, Section{Title: t})
		}
	}
	r.Ingredients = dropEmptySections(r.Ingredients)
	r.Steps = dropEmptySections(r.Steps)

	r.Pages = pagesIn(u)
	p.gate(&r, u.source())
	return r
}

// gate is the verification layer: every verbatim string must be a literal
// substring of the source, plus structural invariants.
func (p *Profile) gate(r *Recipe, source string) {
	contains := func(label, v string) {
		if v == "" {
			return
		}
		if !strings.Contains(source, normalize(v)) {
			r.Issues = append(r.Issues, label+" not found verbatim in source: "+truncate(v))
		}
	}
	contains("title", r.VerbatimTitle)
	for _, s := range r.Ingredients {
		contains("ingredient group", s.Title)
		for _, it := range s.Items {
			contains("ingredient", it)
		}
	}
	for _, s := range r.Steps {
		contains("step group", s.Title)
		for _, it := range s.Items {
			contains("step", it)
		}
	}
	if r.Title == "" {
		r.Issues = append(r.Issues, "empty title")
	}
	if r.IngredientCount() == 0 {
		r.Issues = append(r.Issues, "no ingredients")
	}
	if r.StepCount() == 0 {
		r.Issues = append(r.Issues, "no steps")
	}
}

func truncate(s string) string {
	if len([]rune(s)) <= 60 {
		return s
	}
	return string([]rune(s)[:57]) + "..."
}

// Extract applies the profile to the book, capturing one Recipe per well-formed
// unit and running every recipe through the verification gate.
func (p *Profile) Extract(docs []Document) Report {
	units := filter(segment(docs, p.Unit), wellFormed)
	recipes := make([]Recipe, 0, len(units))
	for _, u := range units {
		recipes = append(recipes, p.buildRecipe(u))
	}
	p.assignImages(docs, units, recipes)

	var rep Report
	ok := 0
	for _, r := range recipes {
		if r.OK() {
			ok++
		} else {
			rep.Flagged++
		}
	}
	rep.Recipes = recipes
	overall := 0.0
	if len(units) > 0 {
		overall = float64(ok) / float64(len(units))
	}
	rep.Confidence = Confidence{
		PerField: fieldCoverage(units, p.Fields),
		Overall:  overall,
		NRecipes: len(units),
	}
	return rep
}
