package mela_test

import (
	"image"
	"image/color"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jphastings/mela-recipes"
)

func TestParseRecipe(t *testing.T) {
	f, err := os.Open("fixtures/a.melarecipe")
	if err != nil {
		t.Error(err)
		return
	}

	recipe, err := mela.ParseRecipe(f)
	if err != nil {
		t.Error(err)
		return
	}

	EnsureRecipe(t, recipe, err, "a")
}

func TestParseRecipes(t *testing.T) {
	f, err := os.Open("fixtures/a+b.melarecipes")
	if err != nil {
		t.Error(err)
		return
	}

	fs, err := f.Stat()
	if err != nil {
		t.Error(err)
		return
	}

	i := 0
	expectedIDs := []string{"b", "a"}

	checkRecipes := func(recipe mela.Recipe, err error) {
		expectedID := expectedIDs[i]
		i++
		EnsureRecipe(t, recipe, err, expectedID)
	}

	if err := mela.ParseRecipes(f, fs.Size(), checkRecipes); err != nil {
		t.Error(err)
		return
	}
}

var fixtures = map[string]mela.RawRecipe{
	"a": {
		RawTitle:        "A title",
		RawCategories:   []string{"a", "aa", "aaa"},
		RawYield:        "1",
		RawLink:         "https://example.com/a",
		RawIngredients:  "A ingredients",
		RawText:         "A text",
		RawPrepTime:     "1m", // time.Duration notation here
		RawCookTime:     "1h", // time.Duration notation here
		RawInstructions: "A instructions",
		RawNutrition:    "A nutrition",
	},
	"b": {
		RawTitle:        "B title",
		RawCategories:   []string{"b", "bb"},
		RawYield:        "2",
		RawLink:         "https://example.com/b",
		RawIngredients:  "B ingredients",
		RawText:         "B text",
		RawPrepTime:     "2m", // time.Duration notation here
		RawCookTime:     "2h", // time.Duration notation here
		RawInstructions: "B instructions",
		RawNutrition:    "B nutrition",
	},
}

func EnsureRecipe(t *testing.T, got mela.Recipe, err error, wantID string) {
	if err != nil {
		t.Error(err)
		return
	}

	if got.ID() != wantID {
		t.Errorf("Incorrect Recipe ID: want = %s, got = %s", wantID, got.ID())
		return
	}

	want, ok := fixtures[wantID]
	if !ok {
		// Only test deep if there's a fixture for it
		return
	}

	// Simple Fields

	if got.Title() != want.Title() {
		t.Errorf("For %s, incorrect Recipe Title: want = %s, got = %s", wantID, want.Title(), got.Title())
	}

	if got.Link() != want.Link() {
		t.Errorf("For %s, incorrect Recipe Link: want = %s, got = %s", wantID, want.Link(), got.Link())
	}

	if got.Text() != want.Text() {
		t.Errorf("For %s, incorrect Recipe Text: want = %s, got = %s", wantID, want.Text(), got.Text())
	}

	if got.Ingredients() != want.Ingredients() {
		t.Errorf("For %s, incorrect Recipe Ingredients: want = %s, got = %s", wantID, want.Ingredients(), got.Ingredients())
	}

	if got.Instructions() != want.Instructions() {
		t.Errorf("For %s, incorrect Recipe Instructions: want = %s, got = %s", wantID, want.Instructions(), got.Instructions())
	}

	if got.Nutrition() != want.Nutrition() {
		t.Errorf("For %s, incorrect Recipe Nutrition: want = %s, got = %s", wantID, want.Nutrition(), got.Nutrition())
	}

	if !reflect.DeepEqual(got.Categories(), want.Categories()) {
		t.Errorf("For %s, incorrect Recipe Categories: want = %s, got = %s", wantID, want.Categories(), got.Categories())
	}

	// Images
	// Assumption: All the test images are single pixel transparent PNGs
	imgCount := 0
	got.Images(func(img image.Image, err error) {
		imgCount++

		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(img.At(0, 0), color.NRGBA{}) {
			t.Errorf("For %s, Recipe Image not reference 1px x 1px transparent", wantID)
		}
	})

	if imgCount != 1 {
		t.Errorf("For %s, Recipe Image count incorrect: want = 1, got = %d", wantID, imgCount)
	}

	// Simple parsers: Yield

	wantYield, wantErr := want.Yield()
	if wantErr != nil {
		t.Errorf("For %s, yield fixture incorrect: %v", wantID, err)
		return
	}

	gotYield, err := got.Yield()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe Yield: %v", wantID, err)
	} else if gotYield != wantYield {
		t.Errorf("For %s, incorrect Recipe Yield: want = %d, got = %d", wantID, wantYield, gotYield)
	}

	// Simple parsers: PrepTime

	wantPrepTime, wantErr := want.PrepTime()
	if wantErr != nil {
		t.Errorf("For %s, PrepTime fixture incorrect: %v", wantID, err)
		return
	}

	gotPrepTime, err := got.PrepTime()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe PrepTime: %v", wantID, err)
	} else if durationsSame(wantPrepTime, gotPrepTime) {
		t.Errorf("For %s, incorrect Recipe PrepTime: want = %v, got = %v", wantID, wantPrepTime, gotPrepTime)
	}

	// Simple parsers: CookTime

	wantCookTime, wantErr := want.CookTime()
	if wantErr != nil {
		t.Errorf("For %s, CookTime fixture incorrect: %v", wantID, err)
		return
	}

	gotCookTime, err := got.CookTime()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe CookTime: %v", wantID, err)
	} else if durationsSame(wantCookTime, gotCookTime) {
		t.Errorf("For %s, incorrect Recipe CookTime: want = %v, got = %v", wantID, wantCookTime, gotCookTime)
	}

	// Simple parsers: TotalTime

	wantTotalTime, wantErr := want.TotalTime()
	if wantErr != nil {
		t.Errorf("For %s, TotalTime fixture incorrect: %v", wantID, err)
		return
	}

	gotTotalTime, err := got.TotalTime()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe TotalTime: %v", wantID, err)
	} else if durationsSame(wantTotalTime, gotTotalTime) {
		t.Errorf("For %s, incorrect Recipe TotalTime: want = %v, got = %v", wantID, wantTotalTime, gotTotalTime)
	}
}

func durationsSame(want, got *time.Duration) bool {
	return (want == nil && got != nil) || (want != nil && *got != *want)
}
