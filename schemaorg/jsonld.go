package schemaorg

import (
	"encoding/json"
	"strings"

	"golang.org/x/net/html"
)

// jsonLDRecipe finds the first schema.org Recipe node across every
// <script type="application/ld+json"> block in the document.
func jsonLDRecipe(doc *html.Node) (map[string]any, bool) {
	for _, block := range jsonLDBlocks(doc) {
		var v any
		if err := json.Unmarshal([]byte(block), &v); err != nil {
			continue
		}
		if node, ok := findRecipeNode(v); ok {
			return node, true
		}
	}
	return nil, false
}

func jsonLDBlocks(doc *html.Node) []string {
	var blocks []string
	forEach(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" &&
			strings.EqualFold(strings.TrimSpace(getAttr(n, "type")), "application/ld+json") {
			blocks = append(blocks, nodeRawText(n))
		}
	})
	return blocks
}

// nodeRawText concatenates a node's raw text children. Inside <script> the parser
// keeps content verbatim (entities undecoded), which is what JSON needs.
func nodeRawText(n *html.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	}
	return sb.String()
}

// findRecipeNode walks a decoded JSON-LD value (object, array, or a
// {"@graph":[…]} wrapper) and returns the first node whose @type is, or contains,
// "Recipe".
func findRecipeNode(v any) (map[string]any, bool) {
	switch t := v.(type) {
	case []any:
		for _, e := range t {
			if n, ok := findRecipeNode(e); ok {
				return n, true
			}
		}
	case map[string]any:
		if isType(t["@type"], "Recipe") {
			return t, true
		}
		if g, ok := t["@graph"]; ok {
			if n, ok := findRecipeNode(g); ok {
				return n, true
			}
		}
	}
	return nil, false
}

// isType reports whether a JSON-LD @type value (a string or array of strings)
// names the given type, tolerating namespace prefixes and URLs (eg.
// "schema:Recipe", "http://schema.org/Recipe").
func isType(v any, want string) bool {
	match := func(s string) bool {
		s = strings.TrimSpace(s)
		return strings.EqualFold(s, want) ||
			strings.HasSuffix(s, "/"+want) ||
			strings.HasSuffix(s, ":"+want)
	}
	switch t := v.(type) {
	case string:
		return match(t)
	case []any:
		for _, e := range t {
			if s, ok := e.(string); ok && match(s) {
				return true
			}
		}
	}
	return false
}
