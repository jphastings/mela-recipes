package schemaorg

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// getAttr returns the value of the named attribute, or "" if absent.
func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// classes returns the element's class tokens.
func classes(n *html.Node) []string {
	return strings.Fields(getAttr(n, "class"))
}

// hasClass reports whether the element carries the given class token.
func hasClass(n *html.Node, class string) bool {
	for _, c := range classes(n) {
		if c == class {
			return true
		}
	}
	return false
}

// findFirst returns the first node (depth-first, self included) matching pred.
func findFirst(n *html.Node, pred func(*html.Node) bool) *html.Node {
	if pred(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if f := findFirst(c, pred); f != nil {
			return f
		}
	}
	return nil
}

// forEach visits n and every descendant, depth-first.
func forEach(n *html.Node, fn func(*html.Node)) {
	fn(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEach(c, fn)
	}
}

var blockElements = map[string]bool{
	"p": true, "br": true, "li": true, "div": true, "tr": true,
	"ol": true, "ul": true, "section": true, "article": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
}

// writeText appends a node's visible text, forcing a newline around block-level
// elements so the result can be split into discrete items.
func writeText(n *html.Node, sb *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		sb.WriteString(n.Data)
		return
	case html.ElementNode:
		if n.Data == "script" || n.Data == "style" {
			return
		}
		if n.Data == "br" {
			sb.WriteByte('\n')
			return
		}
	}

	block := n.Type == html.ElementNode && blockElements[n.Data]
	if block {
		sb.WriteByte('\n')
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		writeText(c, sb)
	}
	if block {
		sb.WriteByte('\n')
	}
}

// nodeText returns the visible text of a node with block elements newline-separated.
func nodeText(n *html.Node) string {
	var sb strings.Builder
	writeText(n, &sb)
	return sb.String()
}

var spaces = regexp.MustCompile(`[ \t\f\r]+`)

// cleanText collapses a raw text blob to a single trimmed line.
func cleanText(raw string) string {
	return strings.TrimSpace(spaces.ReplaceAllString(strings.ReplaceAll(raw, "\n", " "), " "))
}

// cleanLines splits a raw text blob into trimmed, non-empty lines.
func cleanLines(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		if l := strings.TrimSpace(spaces.ReplaceAllString(line, " ")); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// parseFragmentText parses an HTML fragment and returns its block-aware text.
func parseFragmentText(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	return nodeText(doc)
}

// htmlText reduces a possibly-HTML string to one line of plain text, decoding
// entities and stripping tags.
func htmlText(s string) string {
	if strings.ContainsAny(s, "<&") {
		s = parseFragmentText(s)
	}
	return cleanText(s)
}

// htmlLines reduces a possibly-HTML string to trimmed, non-empty lines, with
// block elements (and bare newlines) becoming separate lines.
func htmlLines(s string) []string {
	if strings.ContainsAny(s, "<&") {
		s = parseFragmentText(s)
	}
	return cleanLines(s)
}
