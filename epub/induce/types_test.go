package induce

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestSelMatchesEmptyClassIsClasslessOnly(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(
		`<body><p class="method">step</p><p>bare</p><div class="recipe">x</div></body>`))
	if err != nil {
		t.Fatal(err)
	}
	var classed, bare, div *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch {
			case n.Data == "p" && classOf(n) == "method":
				classed = n
			case n.Data == "p" && classOf(n) == "":
				bare = n
			case n.Data == "div":
				div = n
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	// A classless selector must match only the classless paragraph, not every <p>.
	classless := Sel{Tag: "p"}
	if !classless.Matches(bare) {
		t.Error("classless p selector should match a classless <p>")
	}
	if classless.Matches(classed) {
		t.Error("classless p selector must NOT match a classed <p> (this was the //p over-match bug)")
	}

	// A tag+class selector matches only that exact class.
	method := Sel{Tag: "p", Class: "method"}
	if !method.Matches(classed) || method.Matches(bare) {
		t.Error("p.method should match only p.method")
	}
	if !(Sel{Tag: "div", Class: "recipe"}).Matches(div) {
		t.Error("div.recipe should match the container")
	}
}
