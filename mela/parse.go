package mela

import (
	"fmt"
	"os"
	"path"

	"github.com/jphastings/recipes/internal/formats"
)

const ZipFileMagicBytes = "PK\x03\x04"

func Parse(b formats.Bundle, _ formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	filename := b[0]
	ext := path.Ext(filename)

	switch ext {
	case recipeExt:
		pe, err := ParseRecipeFile(filename)
		return pe, nil, err
	case collectionExt:
		pe, cd, err := ParseRecipesFile(filename)
		return pe, cd, err
	default:
		return nil, nil, fmt.Errorf("doesn't appear to be a Mela recipe or collection file")
	}
}

func ParseRecipeFile(filename string) (<-chan formats.ParseEvent, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	// Don't defer close; it needs to remain open as long as the goroutine is live

	pe := make(chan formats.ParseEvent)
	go func(pe chan formats.ParseEvent, f *os.File) {
		r, err := ParseRecipeStream(f)
		if err == nil {
			r.filename = withoutExt(filename)
		}
		f.Close()

		pe <- formats.ParseEvent{Recipe: r, Err: err, I: 1, N: 1}
		close(pe)
	}(pe, f)

	return pe, nil
}
