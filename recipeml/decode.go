package recipeml

import (
	"errors"
	"html"
	"regexp"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
)

// recipe and its children mirror the RecipeML element tree closely enough to map
// into the interchange format. Inline markup inside <step> is captured raw and
// flattened to text.
type recipe struct {
	Head        head       `xml:"head"`
	Description string     `xml:"description"`
	Ingredients ingredient `xml:"ingredients"`
	Directions  directions `xml:"directions"`
	Note        string     `xml:"note"`
}

type head struct {
	Title      string   `xml:"title"`
	Categories []string `xml:"categories>cat"`
	Yield      yield    `xml:"yield"`
}

// yield holds both the quantities from <range><q>…</q></range> and any trailing
// text (eg. a unit like "cookies") that sits directly inside <yield>.
type yield struct {
	Text string   `xml:",chardata"`
	Qtys []string `xml:"range>q"`
}

type ingredient struct {
	Divs   []ingDiv `xml:"ing-div"`
	Direct []ing    `xml:"ing"`
}

type ingDiv struct {
	Title string `xml:"title,attr"`
	Ings  []ing  `xml:"ing"`
}

type ing struct {
	Amts     []amt  `xml:"amt"`
	Item     string `xml:"item"`
	Prep     string `xml:"prep"`
	Modifier string `xml:"modifier"`
}

type amt struct {
	Qty  string `xml:"qty"`
	Unit string `xml:"unit"`
}

type directions struct {
	Steps []step `xml:"step"`
}

type step struct {
	Inner string `xml:",innerxml"`
}

// toInterchange maps one RecipeML <recipe> into the interchange format.
func toInterchange(r recipe) (formats.InterchangeRecipe, error) {
	ir := formats.NewInterchangeRecipe()
	ir.Title = cleanText(r.Head.Title)
	ir.Description = cleanText(r.Description)
	ir.Notes = cleanText(r.Note)
	ir.Yield = reconstructYield(r.Head.Yield)

	for _, c := range r.Head.Categories {
		if t := cleanText(c); t != "" {
			ir.Tags = append(ir.Tags, t)
		}
	}

	ir.Ingredients = mapIngredients(r.Ingredients)

	var steps []string
	for _, s := range r.Directions.Steps {
		if t := flattenText(s.Inner); t != "" {
			steps = append(steps, t)
		}
	}
	if len(steps) > 0 {
		ir.Instructions = []formats.TitledList{{List: steps}}
	}

	if ir.Title == "" {
		return formats.InterchangeRecipe{}, errors.New("RecipeML recipe has no title")
	}
	return ir, nil
}

// mapIngredients turns the RecipeML ingredient tree into structured groups: any
// ungrouped <ing>s form a leading untitled group, then each <ing-div> becomes its
// own titled group.
func mapIngredients(in ingredient) []formats.IngredientGroup {
	var groups []formats.IngredientGroup

	if g := ingredientGroup("", in.Direct); len(g.Items) > 0 {
		groups = append(groups, g)
	}
	for _, d := range in.Divs {
		g := ingredientGroup(cleanText(d.Title), d.Ings)
		if len(g.Items) > 0 || g.Title != "" {
			groups = append(groups, g)
		}
	}
	return groups
}

func ingredientGroup(title string, ings []ing) formats.IngredientGroup {
	g := formats.IngredientGroup{Title: title}
	for i, in := range ings {
		g.Items = append(g.Items, buildIngredient(in, i))
	}
	return g
}

// buildIngredient reconstructs "qty unit item (prep)" from the separate RecipeML
// nodes and parses it into a structured ingredient. The grammar reads ASCII
// quantities like "1/2" and "1 1/2" directly.
func buildIngredient(in ing, order int) ingredients.IngredientUse {
	var parts []string
	if len(in.Amts) > 0 {
		if a := cleanText(in.Amts[0].Qty); a != "" {
			parts = append(parts, a)
		}
		if u := cleanText(in.Amts[0].Unit); u != "" {
			parts = append(parts, u)
		}
	}
	parts = append(parts, cleanText(in.Item))
	line := strings.TrimSpace(strings.Join(parts, " "))

	if note := combinePrep(in.Prep, in.Modifier); note != "" {
		line += " (" + note + ")"
	}
	return ingredients.ParseOrItem(line, order)
}

func combinePrep(prep, modifier string) string {
	var parts []string
	if p := cleanText(prep); p != "" {
		parts = append(parts, p)
	}
	if m := cleanText(modifier); m != "" {
		parts = append(parts, m)
	}
	return strings.Join(parts, ", ")
}

// reconstructYield rebuilds a yield string from the <range> quantities and any
// trailing unit text (eg. "24 cookies", "2-3 servings", or a bare "2 dozen").
func reconstructYield(y yield) string {
	var qtys []string
	for _, q := range y.Qtys {
		if t := cleanText(q); t != "" {
			qtys = append(qtys, t)
		}
	}
	qty := strings.Join(qtys, "-")
	text := cleanText(y.Text)

	switch {
	case qty != "" && text != "":
		return qty + " " + text
	case qty != "":
		return qty
	default:
		return text
	}
}

var tagRE = regexp.MustCompile(`<[^>]*>`)

// flattenText strips any inline markup from raw inner XML and collapses
// whitespace, turning a wrapped/marked-up <step> into a single readable line.
func flattenText(inner string) string {
	return cleanText(html.UnescapeString(tagRE.ReplaceAllString(inner, "")))
}

// cleanText trims and collapses internal whitespace in element text.
func cleanText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
