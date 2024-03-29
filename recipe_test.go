package mela_test

import (
	"bytes"
	"image"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jphastings/mela-recipes"
)

func TestParseRecipe(t *testing.T) {
	for _, fixtureNum := range []string{"a", "b", "c"} {
		f, err := os.Open("fixtures/" + fixtureNum + ".melarecipe")
		if err != nil {
			t.Error(err)
			return
		}

		recipe, err := mela.ParseRecipe(f)
		if err != nil {
			t.Error(err)
			return
		}

		EnsureRecipe(t, recipe, fixtureNum)
	}
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

	checkRecipes := func(recipe *mela.Recipe, err error) {
		if err != nil {
			t.Error(err)
			return
		}

		expectedID := expectedIDs[i]
		i++

		EnsureRecipe(t, recipe, expectedID)
	}

	if err := mela.ParseRecipes(f, fs.Size(), checkRecipes); err != nil {
		t.Error(err)
		return
	}
}

var oneMin = time.Minute
var oneHour = time.Hour
var twoMin = 2 * time.Minute
var twoHour = 2 * time.Hour
var thirtyMin = 30 * time.Minute
var threeHour = 3 * time.Hour

var wantFixtures = map[string]struct {
	ID         string
	Title      string
	Link       string
	Text       string
	Nutrition  string
	Categories []string
	Notes      string

	ParsedIngredients  map[string][]string
	ParsedInstructions map[string][]string
	ParsedYield        uint64
	ParsedPrepTime     *time.Duration
	ParsedCookTime     *time.Duration
	ParsedTotalTime    *time.Duration
}{
	"a": {
		ID:         "a",
		Title:      "A title",
		Categories: []string{"a", "aa", "aaa"},
		Link:       "https://example.com/a",
		Text:       "A text",
		Nutrition:  "A nutrition",
		Notes:      "A notes",

		ParsedYield:        1,
		ParsedIngredients:  map[string][]string{"": {"A ingredients"}},
		ParsedPrepTime:     &oneMin,
		ParsedCookTime:     &oneHour,
		ParsedInstructions: map[string][]string{"": {"A instructions"}},
	},
	"b": {
		ID:         "b",
		Title:      "B title",
		Categories: []string{"b", "bb"},
		Link:       "https://example.com/b",
		Text:       "B text",
		Nutrition:  "B nutrition",
		Notes:      "B notes",

		ParsedYield:        2,
		ParsedIngredients:  map[string][]string{"": {"B ingredients"}},
		ParsedPrepTime:     &twoMin,
		ParsedCookTime:     &twoHour,
		ParsedInstructions: map[string][]string{"": {"B instructions"}},
	},
	"c": {
		ID:         "urn:isbn:9780714863603#pages=42&recipe=3",
		Title:      "C title",
		Categories: []string{"c", "cc"},
		Link:       "Fresh & Easy",
		Text:       "C text",
		Nutrition:  "C nutrition",
		Notes:      "C Notes\n\n_9780714863603, p.42, 3rd_",

		ParsedYield:        3,
		ParsedIngredients:  map[string][]string{"": {"C ingredients"}},
		ParsedPrepTime:     &threeHour,
		ParsedCookTime:     &thirtyMin,
		ParsedInstructions: map[string][]string{"": {"C instructions"}},
	},
}

func EnsureRecipe(t *testing.T, got *mela.Recipe, wantID string) {
	if err := got.Standardize(false); err != nil {
		t.Errorf("For %s, was unable to standardize: %#v", wantID, err)
		return
	}

	want, ok := wantFixtures[wantID]
	if !ok {
		// Only test deep if there's a fixture for it
		return
	}

	// Simple Fields

	if got.ID != want.ID {
		t.Errorf("For %s, incorrect ID: want = %s, got = %s", wantID, want.Title, got.Title)
	}

	if got.Title != want.Title {
		t.Errorf("For %s, incorrect Recipe Title: want = %s, got = %s", wantID, want.Title, got.Title)
	}

	if got.Link != want.Link {
		t.Errorf("For %s, incorrect Recipe Link: want = %s, got = %s", wantID, want.Link, got.Link)
	}

	if got.Text != want.Text {
		t.Errorf("For %s, incorrect Recipe Text: want = %s, got = %s", wantID, want.Text, got.Text)
	}

	if got.Nutrition != want.Nutrition {
		t.Errorf("For %s, incorrect Recipe Nutrition: want = %s, got = %s", wantID, want.Nutrition, got.Nutrition)
	}

	if got.Notes != want.Notes {
		t.Errorf("For %s, incorrect Recipe Notes: want = %#v, got = %#v", wantID, want.Notes, got.Notes)
	}

	if !reflect.DeepEqual(got.Categories, want.Categories) {
		t.Errorf("For %s, incorrect Recipe Categories: want = %v, got = %v", wantID, want.Categories, got.Categories)
	}

	// Images

	// All the fixtures have one image, that is a single pixel transparent PNG; converted they
	// should all a single pixel white JPEG
	if len(got.Images) != 1 {
		t.Errorf("For %s, incorrect number of images: want = %d, got = %v", wantID, 1, len(got.Images))
	}
	img, imgType, err := image.Decode(bytes.NewReader(got.Images[0]))
	if err != nil {
		t.Errorf("For %s, could not decode image: %v", wantID, err)
	}
	if imgType != "jpeg" {
		t.Errorf("For %s, wrong imahe type: want = %s, got = %s", wantID, "jpeg", imgType)
	}
	if img.Bounds().Dx() != 1 || img.Bounds().Dy() != 1 {
		t.Errorf("For %s, wrong image size: want = %dx%d, got = %dx%d", wantID, 1, 1, img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Parsed Fields

	gotIngredients := got.Ingredients.Parse()
	if !reflect.DeepEqual(gotIngredients, want.ParsedIngredients) {
		t.Errorf("For %s, incorrect Recipe Ingredients: want = %s, got = %s", wantID, want.ParsedIngredients, gotIngredients)
	}

	gotInstructions := got.Instructions.Parse()
	if !reflect.DeepEqual(gotInstructions, want.ParsedInstructions) {
		t.Errorf("For %s, incorrect Recipe Instructions: want = %s, got = %s", wantID, want.ParsedInstructions, gotInstructions)
	}

	gotYield, err := got.Yield.Parse()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe Yield: %v", wantID, err)
	} else if gotYield != want.ParsedYield {
		t.Errorf("For %s, incorrect Recipe Yield: want = %d, got = %d", wantID, want.ParsedYield, gotYield)
	}

	gotPrepTime, err := got.PrepTime.Parse()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe PrepTime: %v", wantID, err)
	} else if durationsSame(want.ParsedPrepTime, gotPrepTime) {
		t.Errorf("For %s, incorrect Recipe PrepTime: want = %v, got = %v", wantID, want.ParsedPrepTime, gotPrepTime)
	}

	gotCookTime, err := got.CookTime.Parse()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe CookTime: %v", wantID, err)
	} else if durationsSame(want.ParsedCookTime, gotCookTime) {
		t.Errorf("For %s, incorrect Recipe CookTime: want = %v, got = %v", wantID, want.ParsedCookTime, gotCookTime)
	}

	gotTotalTime, err := got.TotalTime.Parse()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe TotalTime: %v", wantID, err)
	} else if durationsSame(want.ParsedTotalTime, gotTotalTime) {
		t.Errorf("For %s, incorrect Recipe TotalTime: want = %v, got = %v", wantID, want.ParsedTotalTime, gotTotalTime)
	}
}

func durationsSame(want, got *time.Duration) bool {
	return (want == nil && got != nil) || (want != nil && *got != *want)
}
