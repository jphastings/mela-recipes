package mela

import (
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
	Import:              ImportRecipe,
	NewCollection:       NewCollection,
	Parse:               Parse,
	Bundle:              formats.BundleByExtension(recipeExt, collectionExt),
}

func withoutExt(name string) string {
	ext := path.Ext(name)
	return name[0 : len(name)-len(ext)]
}
