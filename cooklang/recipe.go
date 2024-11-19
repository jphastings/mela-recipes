package cooklang

import (
	"fmt"
	"io"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/standardize"
	"github.com/justintout/cooklang-go"
)

var _ formats.Recipe = (*Recipe)(nil)

type Recipe struct {
	r        cooklang.Recipe
	filename string
	images   []io.ReadCloser
	toClose  []io.Closer
}

func (r Recipe) Name() string            { return r.r.Name }
func (r Recipe) Format() *formats.Format { return FormatInfo }
func (r Recipe) Filename() string {
	if r.filename == "" {
		return standardize.StringToFilename(r.r.Name) + FormatInfo.Extension
	}
	return r.filename + FormatInfo.Extension
}

func (r *Recipe) Export() (formats.InterchangeRecipe, error) {
	ir := formats.InterchangeRecipe{
		// filename: r.filename, TODO: Set filename of interchange?
		Title: r.r.Name,

		// TODO: Other attributes
	}

	return ir, nil
}

func (r *Recipe) Marshal(io.Writer) error {
	return fmt.Errorf("marshalling cooklang not yet implemented")
}

func (r *Recipe) Standardize() ([]standardize.Std, error) {
	return nil, fmt.Errorf("export not yet implemented")
}
