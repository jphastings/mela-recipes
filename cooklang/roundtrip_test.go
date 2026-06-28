package cooklang

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
	"github.com/jphastings/recipes/internal/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func durPtr(d time.Duration) *time.Duration { return &d }

func toCook(t *testing.T, ir formats.InterchangeRecipe) string {
	t.Helper()
	var b strings.Builder
	require.NoError(t, marshalCook(ir, &b))
	return b.String()
}

func fromCook(t *testing.T, s string) formats.InterchangeRecipe {
	t.Helper()
	ir, err := parseCook(strings.NewReader(s))
	require.NoError(t, err)
	return ir
}

// irGroup builds an interchange ingredient group from free-text lines.
func irGroup(title string, lines ...string) formats.IngredientGroup {
	g := formats.IngredientGroup{Title: title}
	for i, l := range lines {
		g.Items = append(g.Items, ingredients.ParseOrItem(l, i))
	}
	return g
}

// ingredientLines renders structured groups back to their canonical string lines.
func ingredientLines(groups []formats.IngredientGroup) []formats.TitledList {
	var out []formats.TitledList
	for _, g := range groups {
		tl := formats.TitledList{Title: g.Title}
		for _, iu := range g.Items {
			tl.List = append(tl.List, ingredients.FormatIngredientUse(iu))
		}
		out = append(out, tl)
	}
	return out
}

func canonAmount(a *ingredients.Amount) *ingredients.Amount {
	if a == nil {
		return nil
	}
	r, _ := new(big.Rat).SetString((*big.Rat)(a).RatString())
	return (*ingredients.Amount)(r)
}

// normalizeIR collapses empty-vs-nil slices and zeroes the volatile per-ingredient
// identity (Order, UUIDs) so two content-equal interchange recipes compare equal.
func normalizeIR(ir formats.InterchangeRecipe) formats.InterchangeRecipe {
	if len(ir.Images) == 0 {
		ir.Images = nil
	}
	if len(ir.Tags) == 0 {
		ir.Tags = nil
	}
	if len(ir.Instructions) == 0 {
		ir.Instructions = nil
	}

	if len(ir.Ingredients) == 0 {
		ir.Ingredients = nil
	} else {
		groups := make([]formats.IngredientGroup, len(ir.Ingredients))
		for i, g := range ir.Ingredients {
			items := make([]ingredients.IngredientUse, len(g.Items))
			for j, iu := range g.Items {
				iu.Order = 0
				iu.UUID = uuid.UUID{}
				iu.Ingredient.UUID = uuid.UUID{}
				iu.Quantity.Amount = canonAmount(iu.Quantity.Amount)
				items[j] = iu
			}
			groups[i] = formats.IngredientGroup{Title: g.Title, Items: items}
		}
		ir.Ingredients = groups
	}
	return ir
}

// TestRoundtripDirectionB: a Cooklang-representable interchange recipe survives
// interchange -> .cook -> interchange unchanged. Cases keep ingredient order
// aligned with the order names appear in the steps, which is the representable
// subset (see TestInlineReordersUnmentionedIngredient for the boundary).
func TestRoundtripDirectionB(t *testing.T) {
	cases := map[string]formats.InterchangeRecipe{
		"minimal": {
			Title:        "Bread",
			Ingredients:  []formats.IngredientGroup{irGroup("", "200 g flour")},
			Instructions: []formats.TitledList{{List: []string{"Mix the flour"}}},
		},
		"full metadata": {
			Title:        "Focaccia",
			Description:  "A simple Italian flatbread",
			Source:       formats.Source{Name: "Jane's Kitchen", URI: "https://example.com/focaccia"},
			Yield:        "8 servings",
			Notes:        "Best eaten warm\nKeeps two days",
			Tags:         []string{"bread", "italian"},
			PrepTime:     durPtr(90 * time.Minute),
			CookTime:     durPtr(25 * time.Minute),
			TotalTime:    durPtr(115 * time.Minute),
			Ingredients:  []formats.IngredientGroup{irGroup("", "500 g flour")},
			Instructions: []formats.TitledList{{List: []string{"Knead the flour"}}},
		},
		"units": {
			Title: "Everything",
			Ingredients: []formats.IngredientGroup{irGroup("",
				"2 tbsp oil", "1 tsp vanilla", "1 cup cream", "180 ml wine",
				"11 g salt", "1 kg beef", "1 lb potatoes", "1 oz butter",
				"1 pinch pepper", "1 bunch basil",
			)},
			Instructions: []formats.TitledList{{List: []string{
				"Combine the oil, vanilla, cream, wine, salt, beef, potatoes, butter, pepper and basil",
			}}},
		},
		"fractions": {
			Title: "Fractions",
			Ingredients: []formats.IngredientGroup{irGroup("",
				"3½ carrots", "1.2 kg beef", "⅓ bunch basil",
			)},
			Instructions: []formats.TitledList{{List: []string{
				"Mix the carrots, beef and basil",
			}}},
		},
		"aligned sections": {
			Title: "Layered",
			Ingredients: []formats.IngredientGroup{
				irGroup("Dough", "500 g flour", "7 g yeast"),
				irGroup("Topping", "100 g cheese"),
			},
			Instructions: []formats.TitledList{
				{Title: "Dough", List: []string{"Mix the flour and yeast"}},
				{Title: "Topping", List: []string{"Scatter the cheese"}},
			},
		},
		"standalone fallback keeps order": {
			Title: "Fry up",
			Ingredients: []formats.IngredientGroup{irGroup("",
				"onion", "1 pinch salt",
			)},
			Instructions: []formats.TitledList{{List: []string{"Fry the onion"}}},
		},
		"ingredient note": {
			Title:        "Noted",
			Ingredients:  []formats.IngredientGroup{irGroup("", "200 g flour (sifted)")},
			Instructions: []formats.TitledList{{List: []string{"Add the flour"}}},
		},
		"escaped metacharacters": {
			Title:        "Saucy",
			Ingredients:  []formats.IngredientGroup{irGroup("", "hot@home sauce")},
			Instructions: []formats.TitledList{{List: []string{"Add the hot@home sauce"}}},
		},
		"untitled then titled": {
			Title: "Two groups",
			Ingredients: []formats.IngredientGroup{
				irGroup("", "egg"),
				irGroup("Sauce", "onion"),
			},
			Instructions: []formats.TitledList{
				{List: []string{"Beat the egg"}},
				{Title: "Sauce", List: []string{"Soften the onion"}},
			},
		},
	}

	for name, ir := range cases {
		t.Run(name, func(t *testing.T) {
			got := fromCook(t, toCook(t, ir))
			assert.Equal(t, normalizeIR(ir), normalizeIR(got))
		})
	}
}

// TestMarshalIsByteStable: re-marshalling a parsed document reproduces it exactly,
// so .cook output is idempotent once written.
func TestMarshalIsByteStable(t *testing.T) {
	ir := formats.InterchangeRecipe{
		Title:        "Stable",
		Yield:        "4",
		Tags:         []string{"quick"},
		Ingredients:  []formats.IngredientGroup{irGroup("", "2 tbsp oil", "1 onion")},
		Instructions: []formats.TitledList{{List: []string{"Heat the oil and add the onion"}}},
	}
	once := toCook(t, ir)
	twice := toCook(t, fromCook(t, once))
	assert.Equal(t, once, twice)
}

// TestImportLosses checks the lossiness declaration reports exactly the fields a
// recipe would lose when written to Cooklang. Cooklang always re-renders
// ingredients as inline @markup (normalising wording and ordering them by where
// their names appear in the steps), so any recipe with ingredients reports it.
func TestImportLosses(t *testing.T) {
	withIngredients := formats.InterchangeRecipe{
		Title:        "x",
		Ingredients:  []formats.IngredientGroup{irGroup("", "2 tbsp sugar")},
		Instructions: []formats.TitledList{{List: []string{"Add the sugar"}}},
	}
	var fields []string
	for _, lf := range FormatInfo.ImportLosses(withIngredients) {
		fields = append(fields, lf.Field)
	}
	assert.Equal(t, []string{"Ingredients"}, fields)

	none := formats.InterchangeRecipe{
		Title:        "x",
		Instructions: []formats.TitledList{{List: []string{"Wait"}}},
	}
	assert.Empty(t, FormatInfo.ImportLosses(none))
}

// TestInlineReordersUnmentionedIngredient documents the inline boundary: an
// ingredient whose name does not appear in the step is appended after the ones
// that do, so an interchange order that puts it first is not preserved.
func TestInlineReordersUnmentionedIngredient(t *testing.T) {
	ir := formats.InterchangeRecipe{
		Title:        "x",
		Ingredients:  []formats.IngredientGroup{irGroup("", "1 pinch salt", "onion")},
		Instructions: []formats.TitledList{{List: []string{"Fry the onion"}}},
	}
	got := fromCook(t, toCook(t, ir))
	assert.Equal(t, []formats.TitledList{
		{List: []string{"onion", "1 pinch salt"}},
	}, ingredientLines(got.Ingredients))
}
