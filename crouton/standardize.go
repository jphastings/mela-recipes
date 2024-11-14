package crouton

import "github.com/jphastings/recipes/internal/standardize"

func (r *Recipe) Standardize() ([]standardize.Std, error) {
	var stds []standardize.Std
	var stdApplied bool
	if r.filename, stdApplied = standardize.Filename(r.filename, r.RecipeName); stdApplied {
		stds = append(stds, standardize.StdFilename)
	}
	// TODO: More

	return stds, nil
}
