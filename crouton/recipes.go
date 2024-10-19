package crouton

import (
	"errors"
	"os"
	"path"
)

type Recipes struct {
	dir         string
	defaultName string
}

type RecipesAdder interface {
	Add(...*Recipe) error
	Close() error
}

func NewRecipesBundle(dir, name string) (RecipesAdder, error) {
	bookDir := path.Join(dir, stringToFilename(name))
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		return nil, err
	}

	return &Recipes{
		dir:         dir,
		defaultName: name,
	}, nil
}

func (rs *Recipes) Close() error {
	return nil
}

func (rs *Recipes) Add(recipes ...*Recipe) error {
	var errs error
	for _, recipe := range recipes {
		_, err := recipe.Save(rs.dir)
		errs = errors.Join(errs, err)
	}

	return errs
}
