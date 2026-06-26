package cooklang

import (
	"strings"
	"testing"
	"time"

	"github.com/jphastings/recipes/internal/formats"
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

// normalizeIR collapses empty-vs-nil slices so two content-equal interchange
// recipes compare equal regardless of how they were built.
func normalizeIR(ir formats.InterchangeRecipe) formats.InterchangeRecipe {
	if len(ir.Images) == 0 {
		ir.Images = nil
	}
	if len(ir.Tags) == 0 {
		ir.Tags = nil
	}
	if len(ir.Ingredients) == 0 {
		ir.Ingredients = nil
	}
	if len(ir.Instructions) == 0 {
		ir.Instructions = nil
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
			Ingredients:  []formats.TitledList{{List: []string{"200 g flour"}}},
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
			Ingredients:  []formats.TitledList{{List: []string{"500 g flour"}}},
			Instructions: []formats.TitledList{{List: []string{"Knead the flour"}}},
		},
		"units": {
			Title: "Everything",
			Ingredients: []formats.TitledList{{List: []string{
				"2 tbsp oil", "1 tsp vanilla", "1 cup cream", "180 ml wine",
				"11 g salt", "1 kg beef", "1 lb potatoes", "1 oz butter",
				"1 pinch pepper", "1 bunch basil",
			}}},
			Instructions: []formats.TitledList{{List: []string{
				"Combine the oil, vanilla, cream, wine, salt, beef, potatoes, butter, pepper and basil",
			}}},
		},
		"fractions": {
			Title: "Fractions",
			Ingredients: []formats.TitledList{{List: []string{
				"3½ carrots", "1.2 kg beef", "⅓ bunch basil",
			}}},
			Instructions: []formats.TitledList{{List: []string{
				"Mix the carrots, beef and basil",
			}}},
		},
		"aligned sections": {
			Title: "Layered",
			Ingredients: []formats.TitledList{
				{Title: "Dough", List: []string{"500 g flour", "7 g yeast"}},
				{Title: "Topping", List: []string{"100 g cheese"}},
			},
			Instructions: []formats.TitledList{
				{Title: "Dough", List: []string{"Mix the flour and yeast"}},
				{Title: "Topping", List: []string{"Scatter the cheese"}},
			},
		},
		"standalone fallback keeps order": {
			Title: "Fry up",
			Ingredients: []formats.TitledList{{List: []string{
				"1 onion", "1 pinch salt",
			}}},
			Instructions: []formats.TitledList{{List: []string{"Fry the onion"}}},
		},
		"escaped metacharacters": {
			Title:        "Saucy",
			Ingredients:  []formats.TitledList{{List: []string{"1 a@b sauce"}}},
			Instructions: []formats.TitledList{{List: []string{"Add the a@b sauce"}}},
		},
		"untitled then titled": {
			Title: "Two groups",
			Ingredients: []formats.TitledList{
				{List: []string{"1 egg"}},
				{Title: "Sauce", List: []string{"1 onion"}},
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
		Ingredients:  []formats.TitledList{{List: []string{"2 tbsp oil", "1 onion"}}},
		Instructions: []formats.TitledList{{List: []string{"Heat the oil and add the onion"}}},
	}
	once := toCook(t, ir)
	twice := toCook(t, fromCook(t, once))
	assert.Equal(t, once, twice)
}

// TestImportLosses checks the lossiness declaration reports exactly the fields a
// recipe would lose when written to Cooklang.
func TestImportLosses(t *testing.T) {
	lossy := formats.InterchangeRecipe{
		Title:        "x",
		Ingredients:  []formats.TitledList{{List: []string{"a handful of basil"}}},
		Instructions: []formats.TitledList{{List: []string{"Add the basil"}}},
	}
	var fields []string
	for _, lf := range FormatInfo.ImportLosses(lossy) {
		fields = append(fields, lf.Field)
	}
	assert.Equal(t, []string{"Ingredients"}, fields)

	clean := formats.InterchangeRecipe{
		Title:        "x",
		Ingredients:  []formats.TitledList{{List: []string{"2 tbsp sugar"}}},
		Instructions: []formats.TitledList{{List: []string{"Add the sugar"}}},
	}
	assert.Empty(t, FormatInfo.ImportLosses(clean))
}

// TestInlineReordersUnmentionedIngredient documents the inline boundary: an
// ingredient whose name does not appear in the step is appended after the ones
// that do, so an interchange order that puts it first is not preserved.
func TestInlineReordersUnmentionedIngredient(t *testing.T) {
	ir := formats.InterchangeRecipe{
		Title:        "x",
		Ingredients:  []formats.TitledList{{List: []string{"1 pinch salt", "1 onion"}}},
		Instructions: []formats.TitledList{{List: []string{"Fry the onion"}}},
	}
	got := fromCook(t, toCook(t, ir))
	assert.Equal(t, []formats.TitledList{
		{List: []string{"1 onion", "1 pinch salt"}},
	}, got.Ingredients)
}
