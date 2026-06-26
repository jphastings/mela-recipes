package cooklang

import (
	"errors"
	"io"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/utils"
)

var _ formats.Recipe = (*Recipe)(nil)

// Recipe is a Cooklang recipe. Cooklang is treated as a text projection of the
// interchange format: the parsed/imported interchange recipe is held directly,
// and Marshal renders it back to .cook text.
type Recipe struct {
	filename string
	ir       formats.InterchangeRecipe
}

func (r Recipe) Name() string            { return r.ir.Title }
func (r Recipe) Format() *formats.Format { return FormatInfo }

func (r Recipe) Filename() string {
	if r.filename == "" {
		return standardize.StringToFilename(r.ir.Title) + FormatInfo.Extension
	}
	return r.filename + FormatInfo.Extension
}

func (r *Recipe) Export() (formats.InterchangeRecipe, error) {
	return r.ir, nil
}

func (r *Recipe) Marshal(w io.Writer) error {
	return marshalCook(r.ir, w)
}

func (r *Recipe) Standardize() ([]standardize.Std, error) {
	var stds []standardize.Std
	var errs []error

	var applied bool
	if r.filename, applied = standardize.Filename(r.filename, r.ir.Title); applied {
		stds = append(stds, standardize.StdFilename)
	}

	if r.ir.Images == nil {
		r.ir.Images = make([]utils.B64Image, 0)
	}
	resized, err := utils.OptimizeImages(r.ir.Images)
	if err != nil {
		errs = append(errs, err)
	}
	if resized {
		stds = append(stds, standardize.StdImages)
	}

	return stds, errors.Join(errs...)
}
