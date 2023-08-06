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

		EnsureRecipe(t, recipe.Standardize(), err, fixtureNum)
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

var oneMin = time.Minute
var oneHour = time.Hour
var twoMin = 2 * time.Minute
var twoHour = 2 * time.Hour
var thirtyMin = 30 * time.Minute
var threeHour = 3 * time.Hour

var wantFixtures = map[string]struct {
	ID           string
	Title        string
	Link         string
	Text         string
	Ingredients  []string
	Instructions []string
	Nutrition    string
	Categories   []string
	Notes        string

	Images    []string
	Yield     uint64
	PrepTime  *time.Duration
	CookTime  *time.Duration
	TotalTime *time.Duration
}{
	"a": {
		ID:           "a",
		Title:        "A title",
		Categories:   []string{"a", "aa", "aaa"},
		Yield:        1,
		Link:         "https://example.com/a",
		Ingredients:  []string{"A ingredients"},
		Text:         "A text",
		PrepTime:     &oneMin,
		CookTime:     &oneHour,
		Instructions: []string{"A instructions"},
		Nutrition:    "A nutrition",
		Notes:        "A notes",
	},
	"b": {
		ID:           "b",
		Title:        "B title",
		Categories:   []string{"b", "bb"},
		Yield:        2,
		Link:         "https://example.com/b",
		Ingredients:  []string{"B ingredients"},
		Text:         "B text",
		PrepTime:     &twoMin,
		CookTime:     &twoHour,
		Instructions: []string{"B instructions"},
		Nutrition:    "B nutrition",
		Notes:        "B notes",
	},
	"c": {
		ID:           "urn:isbn:9780198526636#pages=42&recipe=3",
		Title:        "C title",
		Categories:   []string{"c", "cc"},
		Yield:        3,
		Link:         "https://example.com/c",
		Ingredients:  []string{"C ingredients"},
		Text:         "C text",
		PrepTime:     &threeHour,
		CookTime:     &thirtyMin,
		Instructions: []string{"C instructions"},
		Nutrition:    "C nutrition",
		Notes:        "C Notes",
	},
}

func EnsureRecipe(t *testing.T, got mela.Recipe, err error, wantID string) {
	if err != nil {
		t.Error(err)
		return
	}

	want, ok := wantFixtures[wantID]
	if !ok {
		// Only test deep if there's a fixture for it
		return
	}

	// Simple Fields

	if got.ID() != want.ID {
		t.Errorf("For %s, incorrect ID: want = %s, got = %s", wantID, want.Title, got.Title())
	}

	if got.Title() != want.Title {
		t.Errorf("For %s, incorrect Recipe Title: want = %s, got = %s", wantID, want.Title, got.Title())
	}

	if got.Link() != want.Link {
		t.Errorf("For %s, incorrect Recipe Link: want = %s, got = %s", wantID, want.Link, got.Link())
	}

	if got.Text() != want.Text {
		t.Errorf("For %s, incorrect Recipe Text: want = %s, got = %s", wantID, want.Text, got.Text())
	}

	if !reflect.DeepEqual(got.Ingredients(), want.Ingredients) {
		t.Errorf("For %s, incorrect Recipe Ingredients: want = %s, got = %s", wantID, want.Ingredients, got.Ingredients())
	}

	if !reflect.DeepEqual(got.Instructions(), want.Instructions) {
		t.Errorf("For %s, incorrect Recipe Instructions: want = %s, got = %s", wantID, want.Instructions, got.Instructions())
	}

	if got.Nutrition() != want.Nutrition {
		t.Errorf("For %s, incorrect Recipe Nutrition: want = %s, got = %s", wantID, want.Nutrition, got.Nutrition())
	}

	if got.Notes() != want.Notes {
		t.Errorf("For %s, incorrect Recipe Notes: want = %#v, got = %#v", wantID, want.Notes, got.Notes())
	}

	if !reflect.DeepEqual(got.Categories(), want.Categories) {
		t.Errorf("For %s, incorrect Recipe Categories: want = %v, got = %v", wantID, want.Categories, got.Categories())
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

	gotYield, err := got.Yield()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe Yield: %v", wantID, err)
	} else if gotYield != want.Yield {
		t.Errorf("For %s, incorrect Recipe Yield: want = %d, got = %d", wantID, want.Yield, gotYield)
	}

	gotPrepTime, err := got.PrepTime()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe PrepTime: %v", wantID, err)
	} else if durationsSame(want.PrepTime, gotPrepTime) {
		t.Errorf("For %s, incorrect Recipe PrepTime: want = %v, got = %v", wantID, want.PrepTime, gotPrepTime)
	}

	gotCookTime, err := got.CookTime()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe CookTime: %v", wantID, err)
	} else if durationsSame(want.CookTime, gotCookTime) {
		t.Errorf("For %s, incorrect Recipe CookTime: want = %v, got = %v", wantID, want.CookTime, gotCookTime)
	}

	gotTotalTime, err := got.TotalTime()
	if err != nil {
		t.Errorf("For %s, could not parse Recipe TotalTime: %v", wantID, err)
	} else if durationsSame(want.TotalTime, gotTotalTime) {
		t.Errorf("For %s, incorrect Recipe TotalTime: want = %v, got = %v", wantID, want.TotalTime, gotTotalTime)
	}
}

func durationsSame(want, got *time.Duration) bool {
	return (want == nil && got != nil) || (want != nil && *got != *want)
}
