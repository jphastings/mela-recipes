package mela

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/yeka/zip"
)

type Recipes struct {
	zip *zip.Writer
}

type RecipesAdder interface {
	Add(*Recipe) error
	Close() error
}

// RecipeFunc returns the (streaming) result of parsing a recipe out of a .melarecipes file
type RecipeFunc func(*Recipe, error)

// ParseRecipe parses a known .melarecipes collection file into a stream of Recipe-compatible structs, calling the onRecipe func for each, as it is parsed.
func ParseRecipes(r io.ReaderAt, size int64, onRecipe RecipeFunc) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}

	for _, zf := range zr.File {
		rr, err := zf.Open()
		if err != nil {
			onRecipe(nil, err)
		}
		defer rr.Close()

		if recipe, err := ParseRecipe(rr); err != nil {
			onRecipe(nil, err)
		} else {
			recipe.Filename = withoutExt(zf.Name)
			onRecipe(recipe, nil)
		}
	}

	return nil
}

// NewRecipesBundle creates a .melarecipes (zip file) and allows writing new recipes directly to it with .Add().
//
// If protect argument is true then a .protectedrecipes (zip file) will be created, password protecting all recipe files in a way that requires proof of ownership to decrypt.
// Note that if protect is true then all recipes must have the same ISBN, and all are held in memory before data is written (which may become large).
func NewRecipesBundle(dir, name string, protect bool) (RecipesAdder, error) {
	ext := "melarecipes"
	if protect {
		ext = "protectedrecipes"
	}

	filename := path.Join(dir, stringToFilename(name)+"."+ext)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	if protect {
		return newProtectedRecipes(f), nil
	}

	return &Recipes{
		zip: zip.NewWriter(f),
	}, nil
}

func (rs *Recipes) Close() error {
	return rs.zip.Close()
}

func (rs *Recipes) Add(r *Recipe) error {
	w, err := rs.zip.Create(r.Filename + ".melarecipe")
	if err != nil {
		return fmt.Errorf("unable to create recipe file in zip: %w", err)
	}

	if err := json.NewEncoder(w).Encode(r); err != nil {
		return fmt.Errorf("unable to encode recipe JSON into zip: %w", err)
	}

	return nil
}
