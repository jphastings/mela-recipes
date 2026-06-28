package cooklang

import (
	"testing"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/stretchr/testify/assert"
)

// TestMarshalGolden pins the .cook output format: YAML frontmatter, a blank line,
// then steps with ingredients inlined as @markup.
func TestMarshalGolden(t *testing.T) {
	ir := formats.InterchangeRecipe{
		Title:        "Bread",
		Ingredients:  []formats.IngredientGroup{irGroup("", "200 g flour")},
		Instructions: []formats.TitledList{{List: []string{"Mix the flour"}}},
	}

	want := "---\ntitle: Bread\n---\n\nMix the @flour{200%g}\n"
	assert.Equal(t, want, toCook(t, ir))
}

// TestParseFoldsCookwareAndTimers covers the direction-A (hand-authored) reader:
// cookware/timer markup and comments are folded into readable instruction text,
// and "> " lines become notes.
func TestParseFoldsCookwareAndTimers(t *testing.T) {
	src := "---\n" +
		"title: Boiled Egg\n" +
		"---\n\n" +
		"Bring a #pan of @water{500%ml} to the boil.\n\n" +
		"Add the @egg{1} and cook for ~{6%minutes}. -- soft boiled\n\n" +
		"> Use a timer\n"

	got := fromCook(t, src)

	assert.Equal(t, "Boiled Egg", got.Title)
	assert.Equal(t, []formats.TitledList{
		{List: []string{"500 ml water", "1 egg"}},
	}, ingredientLines(got.Ingredients))
	assert.Equal(t, []formats.TitledList{
		{List: []string{
			"Bring a pan of water to the boil.",
			"Add the egg and cook for 6 minutes.",
		}},
	}, got.Instructions)
	assert.Equal(t, "Use a timer", got.Notes)
}

// TestParseDirectionAByteStable: a canonical .cook document survives
// .cook -> interchange -> .cook unchanged.
func TestParseDirectionAByteStable(t *testing.T) {
	src := "---\ntitle: Bread\n---\n\nMix the @flour{200%g}\n"
	assert.Equal(t, src, toCook(t, fromCook(t, src)))
}

// TestExportLosses reports the always-present cook->interchange loss.
func TestExportLosses(t *testing.T) {
	var fields []string
	for _, lf := range FormatInfo.ExportLosses(formats.InterchangeRecipe{}) {
		fields = append(fields, lf.Field)
	}
	assert.Equal(t, []string{"Instructions"}, fields)
}
