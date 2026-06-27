package paprika

import (
	"fmt"
	"io"
	"path"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/utils"
)

func Parse(b formats.Bundle, opts formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	filename := b[0]
	decode := decodeRecipe(opts.AllowNetwork)

	switch path.Ext(filename) {
	case recipeExt:
		pe, err := formats.ParseRecipeFile(filename, decode)
		return pe, nil, err
	case collectionExt:
		return formats.ParseZipCollection(filename, recipeExt, decode)
	default:
		return nil, nil, fmt.Errorf("doesn't appear to be a Paprika recipe or collection file")
	}
}

// ParseRecipeFile parses a single .paprikarecipe file, streaming one ParseEvent.
func ParseRecipeFile(filename string) (<-chan formats.ParseEvent, error) {
	return formats.ParseRecipeFile(filename, decodeRecipe(false))
}

// ParseRecipesFile parses a .paprikarecipes collection (a zip of gzipped
// .paprikarecipe files), streaming its recipes as they are decoded.
func ParseRecipesFile(filename string) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	return formats.ParseZipCollection(filename, recipeExt, decodeRecipe(false))
}

// decodeRecipe returns the per-format codec that decodes one Paprika recipe
// (gzip JSON) and tags it with its filename. When allowNetwork is set and a
// recipe carries only a remote image_url (no embedded photo), that image is
// fetched so it survives conversion to formats that embed image data.
func decodeRecipe(allowNetwork bool) formats.RecipeDecodeFunc {
	return func(r io.Reader, filename string) (formats.Recipe, error) {
		rec, err := ParseRecipeStream(r)
		if err != nil {
			return nil, err
		}
		rec.filename = filename

		if allowNetwork && len(rec.PhotoData) == 0 && rec.ImageURL != "" {
			if img, err := utils.FetchImage(rec.ImageURL); err == nil {
				rec.PhotoData = img
			}
		}

		return rec, nil
	}
}
