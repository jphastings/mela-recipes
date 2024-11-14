package mela

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/yeka/zip"
)

var _ formats.RecipeCollection = (*RecipeCollection)(nil)

type RecipeCollection struct {
	name     string
	filename string
	recipes  []*Recipe
}

// ParseRecipe parses a known .melarecipes collection file into a RecipeCollection compatible struct.
func ParseRecipes(r io.ReaderAt, size int64) (*RecipeCollection, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	rs := &RecipeCollection{}

	for _, zf := range zr.File {
		if !strings.HasSuffix(zf.Name, "."+recipeExt) {
			continue
		}

		rr, err := zf.Open()
		if err != nil {
			return rs, err
		}
		defer rr.Close()

		if recipe, err := ParseRecipe(rr); err != nil {
			return rs, err
		} else {
			recipe.filename = withoutExt(zf.Name)
			rs.recipes = append(rs.recipes, recipe)
		}
	}

	return rs, nil
}

func (rc *RecipeCollection) Filename() string       { return rc.filename + "." + format.ExtensionCollection }
func (rc *RecipeCollection) Format() formats.Format { return format }

func (rc *RecipeCollection) Add(rs ...formats.Recipe) error {
	for _, ir := range rs {
		if r, ok := ir.(*Recipe); ok {
			rc.recipes = append(rc.recipes, r)
		} else {
			// TODO: convert
			return fmt.Errorf("the provided recipe is not of a format that can be stored in a .melarecipes collection")
		}
	}

	return nil
}

func (rs *RecipeCollection) Recipes() []formats.Recipe {
	out := make([]formats.Recipe, len(rs.recipes))
	for i, r := range rs.recipes {
		out[i] = formats.Recipe(r)
	}
	return out
}

func (rc *RecipeCollection) Marshal(w io.Writer) error {
	z := zip.NewWriter(w)
	defer z.Close()

	var errs error
	for _, recipe := range rc.recipes {
		w, err := z.Create(recipe.filename + ".melarecipe")
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("unable to create recipe file in zip: %w", err))
			continue
		}

		if err := recipe.Marshal(w); err != nil {
			errs = errors.Join(errs, fmt.Errorf("unable to encode recipe JSON into zip: %w", err))
		}
	}

	return errs
}

func (rc *RecipeCollection) Name() string {
	// TODO: this needs a fallback
	return rc.name
}
