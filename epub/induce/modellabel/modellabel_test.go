package modellabel

import (
	"context"
	"strings"
	"testing"

	"github.com/jphastings/recipes/epub/induce"
	"golang.org/x/net/html"
)

func pNodes(t *testing.T, frag string) []*html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(frag))
	if err != nil {
		t.Fatal(err)
	}
	var ps []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			ps = append(ps, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return ps
}

// TestLabelCachesRolePerClass is the regression test for the per-candidate
// re-embedding that made the model refiner time out: a class's role is intrinsic
// to its text, so evaluating a second candidate delimiter over the same content
// must not re-embed any class.
func TestLabelCachesRolePerClass(t *testing.T) {
	blocks := pNodes(t, `<html><body>
	  <p class="T">Egg Fried Rice</p>
	  <p class="I">2 cups rice</p><p class="I">1 egg</p><p class="I">soy sauce</p>
	  <p class="S">Fry the rice in a hot wok.</p><p class="S">Add the beaten egg.</p>
	  <p class="T">Plain Rice</p>
	  <p class="I">1 cup rice</p><p class="I">water to cover</p><p class="I">a pinch of salt</p>
	  <p class="S">Boil the rice gently.</p><p class="S">Cover and steam.</p>
	  <p class="T">Congee</p>
	  <p class="I">1 cup rice</p><p class="I">2 litres water</p><p class="I">fresh ginger</p>
	  <p class="S">Simmer for an hour.</p><p class="S">Season and serve.</p>
	</body></html>`)
	units := []induce.Unit{{Blocks: blocks}}

	var embedded []string
	l := &Labeler{
		roleVec:   map[induce.Role][]float32{},
		roleCache: map[induce.Sel]induce.Role{},
	}
	l.embed = func(_ context.Context, texts []string) ([][]float32, error) {
		embedded = append(embedded, texts...)
		out := make([][]float32, len(texts))
		for i := range out {
			out[i] = []float32{0}
		}
		return out, nil
	}

	l.Label(units, induce.UnitSpec{Mode: induce.ModeHeading, Sel: induce.Sel{Tag: "p", Class: "T"}})
	afterFirst := len(embedded)
	if afterFirst == 0 {
		t.Fatal("expected the first candidate to embed some lines")
	}

	l.Label(units, induce.UnitSpec{Mode: induce.ModeHeading, Sel: induce.Sel{Tag: "p", Class: "S"}})
	if extra := len(embedded) - afterFirst; extra != 0 {
		t.Errorf("second candidate re-embedded %d lines; want 0 (roles are cached per class)", extra)
	}
}
