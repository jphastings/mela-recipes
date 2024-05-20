package mela

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path"
)

type Recipes struct {
	zip *zip.Writer
}

// ParseRecipe parses a known .melarecipes collection file into a stream of Recipe-compatible structs, calling the onRecipe func for each, as it is parsed
func ParseRecipes(r io.ReaderAt, size int64, onRecipe func(*Recipe, error)) error {
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
func NewRecipesBundle(dir, name string) (*Recipes, error) {
	filename := path.Join(dir, stringToFilename(name)+".melarecipes")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
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
		return err
	}

	return json.NewEncoder(w).Encode(r)
}
