package crouton

import (
	"math/big"
	"testing"
	"time"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
	"github.com/jphastings/recipes/internal/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func durPtr(d time.Duration) *time.Duration { return &d }

func mustExtract(t *testing.T, line string, order int) ingredients.IngredientUse {
	t.Helper()
	iu, err := ingredients.ExtractIngredient(line, order)
	require.NoError(t, err)
	return iu
}

func mustSection(t *testing.T, name string, order int) ingredients.IngredientUse {
	t.Helper()
	iu, err := ingredients.NewSection(name, order)
	require.NoError(t, err)
	return iu
}

func step(text string) Step    { return mkStep(false, text) }
func section(text string) Step { return mkStep(true, text) }
func mkStep(isSection bool, text string) Step {
	u, _ := uuid.NewUUID("")
	return Step{IsSection: isSection, Step: text, UUID: u}
}

// normalizeCrouton zeroes the volatile identity fields (regenerated UUIDs,
// senderName, ordering, derived filename) and canonicalises ingredient amounts so
// content equality can be asserted with assert.Equal. The recipe UUID (round-trips
// via the interchange ID) and non-section ingredient UUIDs (deterministic from the
// name) are deliberately kept.
func normalizeCrouton(r *Recipe) *Recipe {
	c := *r
	c.filename = ""
	c.SenderName = ""

	c.Ingredients = append([]ingredients.IngredientUse(nil), r.Ingredients...)
	for i := range c.Ingredients {
		c.Ingredients[i].Order = 0
		c.Ingredients[i].UUID = uuid.UUID{}
		c.Ingredients[i].Quantity.Amount = canonAmount(c.Ingredients[i].Quantity.Amount)
		if c.Ingredients[i].Quantity.Type == ingredients.SectionMarker {
			c.Ingredients[i].Ingredient.UUID = uuid.UUID{}
		}
	}
	if len(c.Ingredients) == 0 {
		c.Ingredients = nil
	}

	c.Steps = append(Steps(nil), r.Steps...)
	for i := range c.Steps {
		c.Steps[i].Order = 0
		c.Steps[i].UUID = uuid.UUID{}
	}
	if len(c.Steps) == 0 {
		c.Steps = nil
	}

	if len(c.Images) == 0 {
		c.Images = nil
	}
	return &c
}

// canonAmount re-creates the rational in its canonical reduced form so that
// reflect.DeepEqual sees identical internal big.Int state for amounts that are the
// same exact value (both routes here parse the same canonical strings).
func canonAmount(a *ingredients.Amount) *ingredients.Amount {
	if a == nil {
		return nil
	}
	r, _ := new(big.Rat).SetString((*big.Rat)(a).RatString())
	return (*ingredients.Amount)(r)
}

func normalizeIR(ir formats.InterchangeRecipe) formats.InterchangeRecipe {
	if len(ir.Images) == 0 {
		ir.Images = nil
	}
	if len(ir.Tags) == 0 {
		ir.Tags = nil
	}
	return ir
}

// TestRoundtripDirectionA: a crouton recipe survives crouton -> interchange ->
// crouton unchanged (content-lossless).
func TestRoundtripDirectionA(t *testing.T) {
	rid, err := uuid.NewUUID("direction-a")
	require.NoError(t, err)

	unitLines := []string{
		"1 carrot", "2 tbsp sugar", "1 tsp vanilla", "1 cup flour",
		"180 ml white wine", "1 cl brandy", "1 dl milk", "1 l stock",
		"1 fl oz sherry", "11 g salt", "1 kg beef", "1 lb potatoes",
		"1 oz butter", "1 pinch pepper", "1 bunch basil", "1 bottle wine",
		"1 can tomatoes", "1 packet yeast",
	}
	allUnits := make([]ingredients.IngredientUse, len(unitLines))
	for i, line := range unitLines {
		allUnits[i] = mustExtract(t, line, i)
	}

	cases := map[string]*Recipe{
		"minimal": {
			UUID:        rid,
			RecipeName:  "Toast",
			Ingredients: []ingredients.IngredientUse{mustExtract(t, "1 slice bread", 0)},
			Steps:       Steps{step("Toast the bread")},
		},
		"every unit": {
			UUID:        rid,
			RecipeName:  "Everything",
			Ingredients: allUnits,
			Steps:       Steps{step("Combine")},
		},
		"fractions and decimals": {
			UUID:       rid,
			RecipeName: "Fractions",
			Ingredients: []ingredients.IngredientUse{
				mustExtract(t, "3½ carrots", 0),
				mustExtract(t, "⅓ bunch basil", 1),
				mustExtract(t, "2⅔ packet yeast", 2),
				mustExtract(t, "1.2 kg beef", 3),
			},
			Steps: Steps{step("Mix")},
		},
		"sections": {
			UUID:       rid,
			RecipeName: "Sectioned",
			Ingredients: []ingredients.IngredientUse{
				mustExtract(t, "2 tbsp oil", 0),
				mustSection(t, "Sauce", 1),
				mustExtract(t, "1 onion", 2),
			},
			Steps: Steps{
				step("Heat the oil"),
				section("Sauce"),
				step("Soften the onion"),
			},
		},
		"full metadata": {
			UUID:            rid,
			RecipeName:      "Soup",
			SourceName:      "BBC Good Food",
			WebLink:         "https://example.com/soup",
			Notes:           "Best served hot",
			Serves:          4,
			Duration:        Minutes(35 * time.Minute),
			CookingDuration: Minutes(20 * time.Minute),
			Ingredients:     []ingredients.IngredientUse{mustExtract(t, "1 onion", 0)},
			Steps:           Steps{step("Simmer")},
			Images:          []B64Image{[]byte("\x89PNG fake bytes")},
		},
		"no durations or serves": {
			UUID:        rid,
			RecipeName:  "Plain",
			Ingredients: []ingredients.IngredientUse{mustExtract(t, "1 egg", 0)},
			Steps:       Steps{step("Boil")},
		},
	}

	for name, orig := range cases {
		t.Run(name, func(t *testing.T) {
			ir, err := orig.Export()
			require.NoError(t, err)
			got, err := importRecipe(ir)
			require.NoError(t, err)

			assert.Equal(t, normalizeCrouton(orig), normalizeCrouton(got.(*Recipe)))
		})
	}
}

// TestRoundtripDirectionB: a Crouton-representable interchange recipe survives
// interchange -> crouton -> interchange unchanged.
func TestRoundtripDirectionB(t *testing.T) {
	id, err := uuid.NewUUID("direction-b")
	require.NoError(t, err)

	cases := map[string]formats.InterchangeRecipe{
		"full": {
			ID:     id.String(),
			Title:  "Tomato Soup",
			Source: formats.Source{Name: "BBC", URI: "https://example.com/soup"},
			Notes:  "Serve hot",
			Yield:  "4",
			Ingredients: []formats.TitledList{
				{List: []string{"2 tbsp olive oil", "1 onion"}},
				{Title: "Garnish", List: []string{"1 tbsp basil"}},
			},
			Instructions: []formats.TitledList{
				{List: []string{"Heat the oil"}},
				{Title: "Finish", List: []string{"Add the basil"}},
			},
			CookTime:  durPtr(20 * time.Minute),
			TotalTime: durPtr(35 * time.Minute),
		},
		"isbn source": {
			ID:     id.String(),
			Title:  "Grandma's Pie",
			Source: formats.Source{Name: "Family Cookbook", URI: "urn:isbn:9781234567897"},
			Yield:  "6",
			Ingredients: []formats.TitledList{
				{List: []string{"500 g flour"}},
			},
			Instructions: []formats.TitledList{
				{List: []string{"Bake"}},
			},
		},
	}

	for name, ir := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := importRecipe(ir)
			require.NoError(t, err)
			re, err := got.(*Recipe).Export()
			require.NoError(t, err)

			assert.Equal(t, normalizeIR(ir), normalizeIR(re))
		})
	}
}

// TestImportTruncatesFreeFormYield locks the one documented direction-B loss for
// yields: Crouton keeps only the whole-number serving count.
func TestImportTruncatesFreeFormYield(t *testing.T) {
	ir := formats.InterchangeRecipe{
		Title:        "x",
		Yield:        "4-6 servings",
		Ingredients:  []formats.TitledList{{List: []string{"1 egg"}}},
		Instructions: []formats.TitledList{{List: []string{"Boil"}}},
	}
	got, err := importRecipe(ir)
	require.NoError(t, err)
	re, err := got.(*Recipe).Export()
	require.NoError(t, err)

	assert.Equal(t, "4", re.Yield)
}

// TestUnsectionedItemsAfterSectionMerge documents a structural boundary: crouton's
// flat section markers associate items with the most recent heading, so unsectioned
// items following a titled section (not Crouton-representable) are absorbed into it.
func TestUnsectionedItemsAfterSectionMerge(t *testing.T) {
	ir := formats.InterchangeRecipe{
		Title: "x",
		Ingredients: []formats.TitledList{
			{Title: "Sauce", List: []string{"1 onion"}},
			{List: []string{"1 egg"}},
		},
		Instructions: []formats.TitledList{{List: []string{"Cook"}}},
	}
	got, err := importRecipe(ir)
	require.NoError(t, err)
	re, err := got.(*Recipe).Export()
	require.NoError(t, err)

	assert.Equal(t, []formats.TitledList{
		{Title: "Sauce", List: []string{"1 onion", "1 egg"}},
	}, re.Ingredients)
}

// TestCroutonImportLosses checks the programmatic lossiness declaration reports the
// actual fields a given recipe would lose when imported into Crouton.
func TestCroutonImportLosses(t *testing.T) {
	lossy := formats.InterchangeRecipe{
		Title:       "x",
		Description: "a description",
		Tags:        []string{"vegan"},
		PrepTime:    durPtr(10 * time.Minute),
		Yield:       "4-6 servings",
		Ingredients: []formats.TitledList{{List: []string{"a handful of basil"}}},
	}

	var fields []string
	for _, lf := range FormatInfo.ImportLosses(lossy) {
		fields = append(fields, lf.Field)
	}
	assert.ElementsMatch(t, []string{"Description", "Tags", "PrepTime", "Yield", "Ingredients"}, fields)

	clean := formats.InterchangeRecipe{
		Title:       "x",
		Yield:       "4",
		Ingredients: []formats.TitledList{{List: []string{"2 tbsp sugar"}}},
	}
	assert.Empty(t, FormatInfo.ImportLosses(clean))

	// "0" parses to no usable serving count and exports back as an empty yield.
	zeroYield := formats.InterchangeRecipe{
		Title:       "x",
		Yield:       "0",
		Ingredients: []formats.TitledList{{List: []string{"2 tbsp sugar"}}},
	}
	var zf []string
	for _, lf := range FormatInfo.ImportLosses(zeroYield) {
		zf = append(zf, lf.Field)
	}
	assert.Equal(t, []string{"Yield"}, zf)
}
