package mela

import (
	"fmt"
	"os"
	"path"

	"github.com/jphastings/recipes/internal/formats"
)

const (
	recipeExt     = ".melarecipe"
	collectionExt = ".melarecipes"
)

var FormatInfo = &formats.Format{
	Name: "Mela",
	URL:  "https://mela.recipes",
	Features: formats.Features{
		ParseRecipe:     true,
		WriteRecipe:     true,
		ParseCollection: true,
		WriteCollection: true,
	},
	Extension:           recipeExt,
	ExtensionCollection: collectionExt,
	New:                 importRecipe,
	NewCollection:       importCollection,
	Parse:               Parse,
	Bundle:              formats.BundleByExtension(recipeExt, collectionExt),
}

const ZipFileMagicBytes = "PK\x03\x04"

func Parse(b formats.Bundle, _ formats.ParseOptions) (formats.Recipe, formats.RecipeCollection, error) {
	filename := b[0]
	ext := path.Ext(filename)
	if ext != recipeExt && ext != collectionExt {
		return nil, nil, fmt.Errorf("doesn't appear to be a Mela recipe or collection file")
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	fs, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}

	magic := make([]byte, 4)
	i, err := f.ReadAt(magic, 0)
	if err != nil {
		return nil, nil, err
	}
	if i < 4 {
		return nil, nil, ErrInvalidMelaFile
	}

	if magic[0] == '{' {
		r, err := ParseRecipe(f)
		r.filename = withoutExt(filename)
		return r, nil, err
	}

	if string(magic) != ZipFileMagicBytes {
		return nil, nil, ErrInvalidMelaFile
	}

	rs, err := ParseRecipes(f, fs.Size())
	if err != nil {
		return nil, nil, err
	}
	rs.filename = withoutExt(filename)
	return nil, rs, nil
}

func withoutExt(name string) string {
	ext := path.Ext(name)
	return name[0 : len(name)-len(ext)]
}
