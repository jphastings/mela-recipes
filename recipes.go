package mela

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path"
)

type Recipes struct {
	zip *zip.Writer
}

func NewRecipesBundle(dir, name string) (*Recipes, error) {
	filename := path.Join(dir, stringToFilename(name)+".melarecipes")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Recipes{
		zip: zip.NewWriter(f),
	}, nil
}

func (rs *Recipes) Close() error {
	return rs.zip.Close()
}

func (rs *Recipes) Add(r *Recipe) error {
	w, err := rs.zip.Create(r.Filename + ".melarecipe")
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(r)
}
