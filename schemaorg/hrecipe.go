package schemaorg

import (
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"golang.org/x/net/html"
)

// hrecipeRecipe extracts a recipe marked up with the h-recipe microformat
// (microformats2 `h-recipe` with p-name/p-ingredient/e-instructions/…, or the
// legacy `hrecipe` with fn/ingredient/instructions classes).
func hrecipeRecipe(doc *html.Node) (formats.InterchangeRecipe, bool) {
	root := findFirst(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && (hasClass(n, "h-recipe") || hasClass(n, "hrecipe"))
	})
	if root == nil {
		return formats.InterchangeRecipe{}, false
	}

	byClass := collectByClass(root)
	ir := formats.NewInterchangeRecipe()

	ir.Title = mfFirstText(byClass, "p-name", "fn")
	ir.Description = mfFirstText(byClass, "p-summary", "summary")
	ir.Yield = mfFirstText(byClass, "p-yield", "yield")

	for _, n := range mfNodes(byClass, "p-ingredient", "ingredient") {
		if s := cleanText(nodeText(n)); s != "" {
			ir.Ingredients = appendItem(ir.Ingredients, s)
		}
	}

	if steps := mfInstructionSteps(mfNodes(byClass, "e-instructions", "instructions")); len(steps) > 0 {
		ir.Instructions = []formats.TitledList{{List: steps}}
	}

	ir.TotalTime = parseISODuration(mfFirstRaw(byClass, "dt-duration", "p-duration", "duration"))

	for _, n := range mfNodes(byClass, "p-category", "category") {
		for _, part := range strings.Split(cleanText(nodeText(n)), ",") {
			if p := strings.TrimSpace(part); p != "" {
				ir.Tags = append(ir.Tags, p)
			}
		}
	}
	ir.Tags = dedupe(ir.Tags)

	ir.Source = formats.Source{Name: mfFirstText(byClass, "p-author", "author")}

	return ir, true
}

// collectByClass indexes every descendant element of root by each of its class
// tokens.
func collectByClass(root *html.Node) map[string][]*html.Node {
	m := map[string][]*html.Node{}
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				for _, cls := range classes(c) {
					m[cls] = append(m[cls], c)
				}
			}
			walk(c)
		}
	}
	walk(root)
	return m
}

func mfInstructionSteps(nodes []*html.Node) []string {
	switch len(nodes) {
	case 0:
		return nil
	case 1:
		return cleanLines(nodeText(nodes[0]))
	default:
		var steps []string
		for _, n := range nodes {
			if s := cleanText(nodeText(n)); s != "" {
				steps = append(steps, s)
			}
		}
		return steps
	}
}

// mfNodes returns the nodes for the first class name that has any.
func mfNodes(m map[string][]*html.Node, classNames ...string) []*html.Node {
	for _, c := range classNames {
		if ns := m[c]; len(ns) > 0 {
			return ns
		}
	}
	return nil
}

func mfFirstText(m map[string][]*html.Node, classNames ...string) string {
	if ns := mfNodes(m, classNames...); len(ns) > 0 {
		return cleanText(nodeText(ns[0]))
	}
	return ""
}

// mfFirstRaw reads a machine value (eg. a dt-duration), preferring the
// datetime / title / value attributes over visible text.
func mfFirstRaw(m map[string][]*html.Node, classNames ...string) string {
	ns := mfNodes(m, classNames...)
	if len(ns) == 0 {
		return ""
	}
	n := ns[0]
	for _, attr := range []string{"datetime", "title", "value"} {
		if v := getAttr(n, attr); v != "" {
			return v
		}
	}
	return cleanText(nodeText(n))
}
