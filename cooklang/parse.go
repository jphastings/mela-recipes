package cooklang

import (
	"errors"
	"os"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/justintout/cooklang-go"
)

func Parse(b formats.Bundle, _ formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	pe := make(chan formats.ParseEvent)

	go func(pe chan formats.ParseEvent, b formats.Bundle) {
		pe <- formats.ParseEvent{N: 1}
		r, err := ParseRecipe(b)
		pe <- formats.ParseEvent{Recipe: r, Err: err, I: 1}
		close(pe)
	}(pe, b)

	return pe, nil, nil
}

func ParseRecipe(b formats.Bundle) (formats.Recipe, error) {
	cookFile, imageFiles := b[0], b[1:]
	cr, err := cooklang.ParseFile(cookFile)
	if err != nil {
		return nil, err
	}
	r := &Recipe{r: cr, filename: cookFile}

	for _, imageFile := range imageFiles {
		f, err := os.Open(imageFile)
		if err != nil {
			for _, c := range r.toClose {
				err = errors.Join(err, c.Close())
			}
			return nil, err
		}
		r.images = append(r.images, f)
		r.toClose = append(r.toClose, f)
	}

	return r, nil
}
