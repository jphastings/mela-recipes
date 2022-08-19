package mela

import (
	"archive/zip"
	"encoding/json"
	"io"
)

func ParseRecipe(r io.Reader) (Recipe, error) {
	var recipe RawRecipe

	dec := json.NewDecoder(r)
	err := dec.Decode(&recipe)
	return recipe, err
}

func ParseRecipes(r io.ReaderAt, size int64, onRecipe func(Recipe, error)) error {
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

		onRecipe(ParseRecipe(rr))
	}

	return nil
}
