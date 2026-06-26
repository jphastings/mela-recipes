package paprika

import (
	"fmt"
	"io"
	"path"

	"github.com/jphastings/recipes/internal/formats"
)

func Parse(b formats.Bundle, _ formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	filename := b[0]

	switch path.Ext(filename) {
	case recipeExt:
		pe, err := ParseRecipeFile(filename)
		return pe, nil, err
	case collectionExt:
		return ParseRecipesFile(filename)
	default:
		return nil, nil, fmt.Errorf("doesn't appear to be a Paprika recipe or collection file")
	}
}

// ParseRecipeFile parses a single .paprikarecipe file, streaming one ParseEvent.
func ParseRecipeFile(filename string) (<-chan formats.ParseEvent, error) {
	return formats.ParseRecipeFile(filename, decodeRecipe)
}

// ParseRecipesFile parses a .paprikarecipes collection (a zip of gzipped
// .paprikarecipe files), streaming its recipes as they are decoded.
func ParseRecipesFile(filename string) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	return formats.ParseZipCollection(filename, recipeExt, decodeRecipe)
}

// decodeRecipe decodes one Paprika recipe (gzip JSON) and tags it with its
// filename. It is the per-format codec the shared zip/file readers call.
func decodeRecipe(r io.Reader, filename string) (formats.Recipe, error) {
	rec, err := ParseRecipeStream(r)
	if err != nil {
		return nil, err
	}
	rec.filename = filename
	return rec, nil
}
