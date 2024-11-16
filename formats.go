package recipes

import (
	"path"
	"strings"

	"github.com/jphastings/recipes/cooklang"
	"github.com/jphastings/recipes/crouton"
	"github.com/jphastings/recipes/epub"
	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/mela"
)

func AvailableFormats() []*formats.Format {
	return []*formats.Format{
		mela.FormatInfo,
		crouton.FormatInfo,
		epub.FormatInfo,
		cooklang.FormatInfo,
	}
}

// Attempts to find a suitable format & parser for all the recipe files given. All found recipes are returned in the first argument, the second argument holds the details of the collection if the input files represent *exactly and only one* collection.
// If no collections are represented, or the files represent a collection *and* other recipes or collections, then this second return value will be nil.
func ParseAll(files []string, o formats.ParseOptions) ([]formats.Recipe, formats.RecipeCollection, error) {
	countCollections := 0
	var soloCollection formats.RecipeCollection
	var rs []formats.Recipe

	for _, f := range AvailableFormats() {
		bundles, unused := f.Bundle(files)
		files = unused

		for _, b := range bundles {
			r, rc, err := f.Parse(b, o)
			if err != nil {
				return nil, nil, err
			}

			if r != nil {
				rs = append(rs, r)
			}

			if rc != nil {
				countCollections++
				rs = append(rs, rc.Recipes()...)

				if countCollections == 1 {
					soloCollection = rc
				} else {
					soloCollection = nil
				}
			}
		}
	}

	return rs, soloCollection, nil
}

type AsType string

const (
	AsTypeAny        AsType = "any"
	AsTypeRecipe     AsType = "recipe"
	AsTypeCollection AsType = "collection"
)

func ParseDestination(to string) (overrideFilename string, asType AsType, format *formats.Format) {
	ext := path.Ext(to)
	if ext != "" && ext != to {
		overrideFilename = to
	}

	for _, f := range AvailableFormats() {
		if f.Extension != "" && f.Extension == ext {
			return overrideFilename, AsTypeRecipe, f
		}
		if f.ExtensionCollection != "" && f.ExtensionCollection == ext {
			return overrideFilename, AsTypeCollection, f
		}
		if strings.EqualFold(to, f.Name) {
			return "", AsTypeAny, f
		}
	}

	return "", AsTypeAny, nil
}
