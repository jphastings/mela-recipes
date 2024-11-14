package mela

import (
	"os"
	"path"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
)

const (
	recipeExt     = "melarecipe"
	collectionExt = "melarecipes"
)

var format = formats.Format{
	Name: "Mela",
	URL:  "https://mela.recipes/",
	Features: formats.Features{
		ParseRecipe:     true,
		WriteRecipe:     true,
		ParseCollection: true,
		WriteCollection: true,
	},
	Extension:           recipeExt,
	ExtensionCollection: collectionExt,
	New:                 newFromInterchange,
	NewCollection:       newFromInterchangeCollection,
	Parse:               Parse,
}

func init() {
	formats.Register(format)
}

const ZipFileMagicBytes = "PK\x03\x04"

func Parse(f *os.File) (formats.Recipe, formats.RecipeCollection, error) {
	filename := path.Base(f.Name())
	ext := strings.TrimPrefix(path.Ext(filename), ".")
	if ext != recipeExt && ext != collectionExt {
		return nil, nil, nil
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
	rs.filename = withoutExt(filename)
	return nil, rs, err
}

func withoutExt(name string) string {
	ext := path.Ext(name)
	return name[0 : len(name)-len(ext)]
}
