package mela

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"os"
)

var ErrInvalidMelaFile = errors.New("given file is neither a melarecipe nor a melarecipes file")
var ErrInvalidMelaRecipeFile = errors.New("given file is not a melarecipe file")
var ErrInvalidMelaRecipesFile = errors.New("given file is not a melarecipes file")

// Open is a smart, file-system based function for opening a .melarecipe or .melarecipes file from disk.
func Open(filename string, onRecipe func(Recipe, error)) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	fs, err := f.Stat()
	if err != nil {
		return err
	}

	magic := make([]byte, 4)
	i, err := f.ReadAt(magic, 0)
	if err != nil {
		return err
	}
	if i < 4 {
		return ErrInvalidMelaFile
	}

	if magic[0] == '{' {
		r, err := ParseRecipe(f)
		onRecipe(r, err)
		return nil
	}

	if string(magic) == "PK\x03\x04" {
		return ParseRecipes(f, fs.Size(), onRecipe)
	}

	return ErrInvalidMelaFile
}

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
