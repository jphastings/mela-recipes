package induce

import (
	"fmt"
	"strings"
	"testing"
)

// splitIngredientDoc mirrors books (eg. The Book of Jewish Food) that style the
// first ingredient with its own class (ingredient1) and the rest with another
// (ingredient), inside a div.recipe container.
func splitIngredientDoc(t *testing.T) Document {
	t.Helper()
	recipes := []struct {
		title string
		first string
		rest  []string
	}{
		{"Abricotines", "500 g dried apricots", []string{"75 g pistachios", "icing sugar"}},
		{"Coconut Rocks", "200 g dried coconut", []string{"100 g sugar", "2 eggs"}},
		{"Almond Macaroons", "250 g ground almonds", []string{"200 g sugar", "3 egg whites"}},
		{"Date Balls", "400 g pitted dates", []string{"100 g walnuts", "2 tbsp honey"}},
		{"Semolina Cake", "300 g semolina", []string{"200 g sugar", "150 g butter"}},
		{"Fig Jam", "1 kg fresh figs", []string{"500 g sugar", "1 lemon"}},
	}

	var b strings.Builder
	b.WriteString("<html><body>")
	for _, r := range recipes {
		b.WriteString(`<div class="recipe">`)
		fmt.Fprintf(&b, `<p class="recipe_title">%s</p>`, r.title)
		fmt.Fprintf(&b, `<p class="intro">%s is a sweet treat that we always made for celebrations at home.</p>`, r.title)
		fmt.Fprintf(&b, `<p class="ingredient1">%s</p>`, r.first)
		for _, ing := range r.rest {
			fmt.Fprintf(&b, `<p class="ingredient">%s</p>`, ing)
		}
		for i := 0; i < 3; i++ {
			fmt.Fprintf(&b, `<p class="method">Mix everything together well and shape it, step number %d of the recipe.</p>`, i+1)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString("</body></html>")

	doc, err := ParseDocument("split.html", strings.NewReader(b.String()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return doc
}

func TestSplitFirstIngredientMerged(t *testing.T) {
	docs := []Document{splitIngredientDoc(t)}
	p, err := Induce(docs, BookIdent{})
	if err != nil {
		t.Fatalf("induce: %v", err)
	}

	ing := p.Fields[RoleIngredients]
	if !ing.Matches(elem(t, "p", "ingredient")) {
		t.Errorf("ingredients should match the ingredient run: %v", ing.XPaths())
	}
	if !ing.Matches(elem(t, "p", "ingredient1")) {
		t.Errorf("the styled first ingredient (ingredient1) was not merged into ingredients: %v", ing.XPaths())
	}

	rep := p.Extract(docs)
	r, ok := find(rep.Recipes, "Abricotines")
	if !ok {
		t.Fatal("Abricotines not extracted")
	}
	var items []string
	for _, s := range r.Ingredients {
		items = append(items, s.Items...)
	}
	if len(items) == 0 || items[0] != "500 g dried apricots" {
		t.Errorf("first ingredient missing; got %v", items)
	}
}
