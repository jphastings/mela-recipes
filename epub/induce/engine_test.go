package induce

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// aikLikeDoc builds a document with the same shape as Asma's Indian Kitchen:
// h2.h2a native title, h2.h2b English subtitle with an optional "v" marker,
// p.intro description, p.ing ingredients, p.method steps.
func aikLikeDoc(t *testing.T) Document {
	t.Helper()
	recipes := []struct {
		title, sub string
		veg        bool
		ings       []string
		steps      []string
	}{
		{"Baingan Bharta", "Puréed aubergine", true,
			[]string{"2 large aubergines", "4 tbsp vegetable oil"},
			[]string{"Roast the aubergines until the skin is charred all over and the flesh is cooked through completely."}},
		{"Saag Aloo", "Spiced potatoes", true,
			[]string{"500 g spinach", "3 medium potatoes", "1 tsp cumin seeds"},
			[]string{"Cook the potatoes until tender, then add the spinach and the spices and simmer for a further while."}},
		{"Murgh Rezala", "Aromatic chicken stew", false,
			[]string{"800 g chicken thighs", "2 onions"},
			[]string{"Brown the chicken pieces in batches, then return them all to the pan with the onions and spices."}},
		{"Peela Pulao", "Lemon rice with cashew nuts", true,
			[]string{"300 g basmati rice", "50 g cashew nuts"},
			[]string{"Wash the rice in several changes of cold water until the water runs completely clear before cooking."}},
		{"Shahi Kofta", "Lamb meatballs in a rich gravy", false,
			[]string{"500 g minced lamb", "2 tbsp yoghurt"},
			[]string{"Shape the mince into small balls, then fry them gently until they are browned evenly on all sides."}},
		{"Mattar Pulao", "Rice with peas", true,
			[]string{"300 g basmati rice", "150 g peas"},
			[]string{"Soak the rice for thirty minutes, then drain it and cook it gently with the peas and whole spices."}},
	}

	var b strings.Builder
	b.WriteString("<html><body>")
	for i, r := range recipes {
		fmt.Fprintf(&b, `<h2 class="h2a"><a id="page_%d">%s</a></h2>`, 10+i, r.title)
		marker := ""
		if r.veg {
			marker = ` <span class="underline">v</span>`
		}
		fmt.Fprintf(&b, `<h2 class="h2b">%s%s</h2>`, r.sub, marker)
		fmt.Fprintf(&b, `<p class="intro">%s is a much-loved dish that we cooked often at home when I was a child growing up.</p>`, r.title)
		for _, ing := range r.ings {
			fmt.Fprintf(&b, `<p class="ing">%s</p>`, ing)
		}
		for _, st := range r.steps {
			fmt.Fprintf(&b, `<p class="method">%s</p>`, st)
		}
	}
	b.WriteString("</body></html>")

	doc, err := ParseDocument("aik.html", strings.NewReader(b.String()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return doc
}

func find(rs []Recipe, title string) (Recipe, bool) {
	for _, r := range rs {
		if r.Title == title {
			return r, true
		}
	}
	return Recipe{}, false
}

func TestInduceDiscoversStructure(t *testing.T) {
	p, err := Induce([]Document{aikLikeDoc(t)}, BookIdent{Title: "Test"})
	if err != nil {
		t.Fatalf("induce: %v", err)
	}
	if got := p.Unit.Sel; got != (Sel{Tag: "h2", Class: "h2a"}) {
		t.Errorf("recipe unit = %v, want h2.h2a", got)
	}
	if !p.Fields[RoleIngredients].Matches(elem(t, "p", "ing")) {
		t.Errorf("ingredients selector did not match p.ing: %v", p.Fields[RoleIngredients].XPaths())
	}
	if len(p.Markers) != 1 || p.Markers[0].Category != "Vegetarian" {
		t.Errorf("expected a v->Vegetarian marker, got %+v", p.Markers)
	}
}

func TestExtractIsVerbatimAndInterprets(t *testing.T) {
	docs := []Document{aikLikeDoc(t)}
	p, err := Induce(docs, BookIdent{})
	if err != nil {
		t.Fatalf("induce: %v", err)
	}
	rep := p.Extract(docs)

	saag, ok := find(rep.Recipes, "Saag Aloo")
	if !ok {
		t.Fatalf("Saag Aloo not extracted; got %d recipes", len(rep.Recipes))
	}
	// Verbatim: the title must not be corrupted (the "Saag"->"Sag" failure).
	if saag.Title != "Saag Aloo" {
		t.Errorf("title = %q, want verbatim %q", saag.Title, "Saag Aloo")
	}
	// Interpretation: the "v" marker becomes a category, stripped from the subtitle.
	if !contains(saag.Categories, "Vegetarian") {
		t.Errorf("categories = %v, want Vegetarian", saag.Categories)
	}
	if saag.Subtitle != "Spiced potatoes" {
		t.Errorf("subtitle = %q, want marker stripped to %q", saag.Subtitle, "Spiced potatoes")
	}
	if saag.IngredientCount() != 3 {
		t.Errorf("ingredient count = %d, want 3 (%v)", saag.IngredientCount(), saag.Ingredients)
	}
	// A non-vegetarian recipe gets no category.
	if murgh, ok := find(rep.Recipes, "Murgh Rezala"); ok && len(murgh.Categories) != 0 {
		t.Errorf("Murgh Rezala categories = %v, want none", murgh.Categories)
	}
	if !saag.OK() {
		t.Errorf("Saag Aloo failed the gate: %v", saag.Issues)
	}
}

func TestGateRejectsNonVerbatim(t *testing.T) {
	p := &Profile{}
	source := normalize("Saag Aloo spiced potatoes 500 g spinach")

	good := &Recipe{Title: "Saag Aloo", VerbatimTitle: "Saag Aloo",
		Ingredients: []Section{{Items: []string{"500 g spinach"}}}, Steps: []Section{{Items: []string{"cook"}}}}
	p.gate(good, source)
	for _, iss := range good.Issues {
		if strings.Contains(iss, "title") {
			t.Errorf("verbatim title was wrongly rejected: %v", good.Issues)
		}
	}

	bad := &Recipe{Title: "Sag Aloo", VerbatimTitle: "Sag Aloo",
		Ingredients: []Section{{Items: []string{"500 g spinach"}}}, Steps: []Section{{Items: []string{"cook"}}}}
	p.gate(bad, source)
	if !hasIssueAbout(bad.Issues, "title") {
		t.Errorf("corrupted title %q was not caught by the gate; issues=%v", bad.Title, bad.Issues)
	}
}

func TestNormalizePreservesWordsFoldsWhitespace(t *testing.T) {
	if got := normalize("Purée  d'aubergine v"); got != "Purée d'aubergine v" {
		t.Errorf("normalize = %q", got)
	}
}

// Mirrors Asma's "two version" recipes: the first ingredient and the group
// headings live in p.ing1 (alongside "Serves N"), the rest in p.ing. The
// content-ratio selector must keep p.ing1 so the first ingredient isn't
// dropped, and titled lines must become section headers, not items.
func TestGroupedIngredientsCaptured(t *testing.T) {
	var b strings.Builder
	b.WriteString("<html><body>")
	for i := 0; i < 6; i++ {
		fmt.Fprintf(&b, `<h2 class="h2a">Recipe %d</h2>`, i)
		b.WriteString(`<p class="ing1"><strong>Serves 4</strong></p>`)
		b.WriteString(`<p class="ing1">2 large baking potatoes</p>`)
		b.WriteString(`<p class="ing1"><strong>For the tempering</strong></p>`)
		b.WriteString(`<p class="ing">1 tsp cumin seeds</p>`)
		b.WriteString(`<p class="ing">2 dried red chillis</p>`)
		b.WriteString(`<p class="method">Boil the potatoes until tender, then temper the spices and fold everything together well.</p>`)
	}
	b.WriteString("</body></html>")
	doc, err := ParseDocument("g.html", strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}

	p, err := Induce([]Document{doc}, BookIdent{})
	if err != nil {
		t.Fatalf("induce: %v", err)
	}
	r, ok := find(p.Extract([]Document{doc}).Recipes, "Recipe 0")
	if !ok {
		t.Fatal("recipe not extracted")
	}
	if r.Yield != "4" {
		t.Errorf("yield = %q, want 4", r.Yield)
	}
	if !sectionsContain(r.Ingredients, "2 large baking potatoes") {
		t.Errorf("p.ing1 first ingredient dropped: %+v", r.Ingredients)
	}
	if !hasSectionTitle(r.Ingredients, "For the tempering") {
		t.Errorf("group heading not captured as a section: %+v", r.Ingredients)
	}
	if sectionsContain(r.Ingredients, "For the tempering") {
		t.Errorf("group heading leaked in as an ingredient item: %+v", r.Ingredients)
	}
	if !r.OK() {
		t.Errorf("gate failed: %v", r.Issues)
	}
}

// A repeated non-numeric class (image captions) must NOT be mistaken for
// ingredients; the numeric run wins. This is the language-agnostic signal that
// keeps nutrition/caption cruft out of the ingredient list.
func TestIngredientsAreNumericRuns(t *testing.T) {
	var b strings.Builder
	b.WriteString("<html><body>")
	for i := 0; i < 6; i++ {
		fmt.Fprintf(&b, `<h2 class="h2a">Dish %d</h2>`, i)
		b.WriteString(`<p class="cap">Photographed on a linen cloth</p>`)
		b.WriteString(`<p class="cap">Styling by someone</p>`)
		b.WriteString(`<p class="ing">200 g plain flour</p>`)
		b.WriteString(`<p class="ing">2 large eggs</p>`)
		b.WriteString(`<p class="ing">300 ml whole milk</p>`)
		b.WriteString(`<p class="method">Whisk the flour and eggs together, then add the milk slowly until you have a smooth batter.</p>`)
	}
	b.WriteString("</body></html>")
	doc, err := ParseDocument("n.html", strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	p, err := Induce([]Document{doc}, BookIdent{})
	if err != nil {
		t.Fatalf("induce: %v", err)
	}
	ingXPaths := strings.Join(p.Fields[RoleIngredients].XPaths(), " ")
	if !strings.Contains(ingXPaths, "ing") {
		t.Errorf("ingredients should be the numeric run p.ing, got %q", ingXPaths)
	}
	if strings.Contains(ingXPaths, "cap") {
		t.Errorf("caption run p.cap was wrongly chosen as ingredients: %q", ingXPaths)
	}
}

// Hanging-indent layouts (e.g. Cooking for Geeks) split each ingredient into a
// quantity block and a name block, both emphasised. The quantity column must be
// chosen as ingredients and each quantity merged with its name.
func TestHangingIndentIngredientsMerge(t *testing.T) {
	rows := [][2]string{{"1", "cup plain flour"}, {"2", "large eggs, beaten"}, {"200", "ml whole milk"}}
	var b strings.Builder
	b.WriteString("<html><body>")
	for i := 0; i < 6; i++ {
		fmt.Fprintf(&b, `<h2 class="h2a">Dish %d</h2>`, i)
		for _, r := range rows {
			fmt.Fprintf(&b, `<p class="qty"><strong>%s</strong></p>`, r[0])
			fmt.Fprintf(&b, `<p class="nm"><strong>%s</strong></p>`, r[1])
		}
		b.WriteString(`<p class="method">Whisk everything together into a smooth batter and cook gently until set and golden.</p>`)
	}
	b.WriteString("</body></html>")
	doc, err := ParseDocument("h.html", strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	p, err := Induce([]Document{doc}, BookIdent{})
	if err != nil {
		t.Fatalf("induce: %v", err)
	}
	r, ok := find(p.Extract([]Document{doc}).Recipes, "Dish 0")
	if !ok {
		t.Fatal("recipe not extracted")
	}
	for _, want := range []string{"1 cup plain flour", "2 large eggs, beaten", "200 ml whole milk"} {
		if !sectionsContain(r.Ingredients, want) {
			t.Errorf("missing merged ingredient %q in %+v", want, r.Ingredients)
		}
	}
	if sectionsContain(r.Ingredients, "1") {
		t.Errorf("bare quantity left unmerged: %+v", r.Ingredients)
	}
}

// Images are associated by content-flow containment, so a photo within a
// recipe's markup is its photo even when it prints on the facing page (a lower
// page number belonging to a neighbour), and a role="presentation" decoration
// (chapter art) is dropped rather than mistaken for the food photo.
func TestImageAssociation(t *testing.T) {
	var b strings.Builder
	b.WriteString("<html><body>")
	for i := 0; i < 6; i++ {
		page := 20 + 2*i
		// Recipe 2's food photo prints on the facing page (23) and, as books do,
		// sits in the *previous* recipe's flow, just before Recipe 2's heading —
		// alongside a role="presentation" chapter decoration.
		if i == 2 {
			b.WriteString(`<p class="image"><a id="page_23"><img src="food.jpg"/></a></p>`)
			b.WriteString(`<p class="image"><img role="presentation" src="deco.jpg"/></p>`)
		}
		fmt.Fprintf(&b, `<h2 class="h2a"><a id="page_%d">Recipe %d</a></h2>`, page, i)
		b.WriteString(`<p class="ing">1 cup flour</p><p class="ing">2 eggs</p><p class="ing">200 ml milk</p>`)
		b.WriteString(`<p class="method">Mix and bake until done, then allow to cool completely before serving.</p>`)
	}
	b.WriteString("</body></html>")
	doc, err := ParseDocument("t.html", strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	p, err := Induce([]Document{doc}, BookIdent{})
	if err != nil {
		t.Fatalf("induce: %v", err)
	}
	recs := p.Extract([]Document{doc}).Recipes
	r2, _ := find(recs, "Recipe 2")
	if !contains(r2.Images, "food.jpg") {
		t.Errorf("in-flow photo (facing-page number) not assigned to its recipe: %v", r2.Images)
	}
	if contains(r2.Images, "deco.jpg") {
		t.Errorf(`role="presentation" decoration was not excluded: %v`, r2.Images)
	}
	for _, r := range recs {
		if r.Title != "Recipe 2" && contains(r.Images, "food.jpg") {
			t.Errorf("photo leaked to %q", r.Title)
		}
	}
}

// When each recipe is wrapped in its own container, an image inside that
// container belongs to it by DOM containment even if its printed page number
// points elsewhere — the case where a photo lives on a different page span.
func TestImageDomContainment(t *testing.T) {
	var b strings.Builder
	b.WriteString("<html><body>")
	for i := 0; i < 6; i++ {
		b.WriteString(`<div class="recipe">`)
		fmt.Fprintf(&b, `<p class="recipe_title">Dish %d</p>`, i)
		b.WriteString(`<p class="ingredient">1 cup flour</p><p class="ingredient">2 eggs</p><p class="ingredient">200 ml milk</p>`)
		b.WriteString(`<p class="step">Mix everything together thoroughly and bake until risen, golden and cooked.</p>`)
		fmt.Fprintf(&b, `<p class="img"><a id="page_%d"><img src="dish%d.jpg"/></a></p>`, 900+i, i) // misleading page
		b.WriteString(`</div>`)
	}
	b.WriteString("</body></html>")
	doc, err := ParseDocument("t.html", strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	p, err := Induce([]Document{doc}, BookIdent{})
	if err != nil {
		t.Fatalf("induce: %v", err)
	}
	recs := p.Extract([]Document{doc}).Recipes
	for i := 0; i < 6; i++ {
		r, ok := find(recs, fmt.Sprintf("Dish %d", i))
		if !ok {
			t.Fatalf("Dish %d missing", i)
		}
		if !contains(r.Images, fmt.Sprintf("dish%d.jpg", i)) {
			t.Errorf("Dish %d should own its photo via DOM containment, got %v", i, r.Images)
		}
		for j := 0; j < 6; j++ {
			if j != i && contains(r.Images, fmt.Sprintf("dish%d.jpg", j)) {
				t.Errorf("Dish %d wrongly grabbed Dish %d's photo", i, j)
			}
		}
	}
}

// helpers

func sectionsContain(secs []Section, item string) bool {
	for _, s := range secs {
		for _, it := range s.Items {
			if it == item {
				return true
			}
		}
	}
	return false
}

func hasSectionTitle(secs []Section, title string) bool {
	for _, s := range secs {
		if s.Title == title {
			return true
		}
	}
	return false
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func hasIssueAbout(issues []string, sub string) bool {
	for _, i := range issues {
		if strings.Contains(i, sub) {
			return true
		}
	}
	return false
}

func elem(t *testing.T, tag, class string) *html.Node {
	t.Helper()
	doc, err := ParseDocument("x", strings.NewReader("<"+tag+" class=\""+class+"\">x</"+tag+">"))
	if err != nil {
		t.Fatal(err)
	}
	var found *html.Node
	eachElement(doc.Root, func(n *html.Node) {
		if n.Data == tag && classOf(n) == class {
			found = n
		}
	})
	return found
}
