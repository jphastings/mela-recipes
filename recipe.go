package mela

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Recipe struct {
	Filename     string            `json:"-"`
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Link         string            `json:"link"`
	Text         string            `json:"text"`
	Ingredients  SectionedSequence `json:"ingredients"`
	Instructions SectionedSequence `json:"instructions"`
	Nutrition    string            `json:"nutrition"`
	Categories   []string          `json:"categories"`
	Notes        string            `json:"notes"`

	Images    []ImageBytes  `json:"images"`
	Yield     PeopleCount   `json:"yield"`
	PrepTime  MaybeDuration `json:"prepTime"`
	CookTime  MaybeDuration `json:"cookTime"`
	TotalTime MaybeDuration `json:"totalTime"`
}

var ErrInvalidMelaFile = errors.New("given file is neither a melarecipe nor a melarecipes file")
var ErrInvalidMelaRecipeFile = errors.New("given file is not a melarecipe file")
var ErrInvalidMelaRecipesFile = errors.New("given file is not a melarecipes file")

const ZipFileMagicBytes = "PK\x03\x04"

// Open is a smart, file-system based function for opening a .melarecipe or .melarecipes file from disk.
// For simplicity's sake, it will silently ignore any invalid recipes within a .melarecipes file, use ParseRecipes for
// greater control.
func Open(filename string) ([]*Recipe, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	fs, err := f.Stat()
	if err != nil {
		return nil, err
	}

	magic := make([]byte, 4)
	i, err := f.ReadAt(magic, 0)
	if err != nil {
		return nil, err
	}
	if i < 4 {
		return nil, ErrInvalidMelaFile
	}

	if magic[0] == '{' {
		r, err := ParseRecipe(f)
		r.Filename = withoutExt(filename)
		return []*Recipe{r}, err
	}

	if string(magic) != ZipFileMagicBytes {
		return nil, ErrInvalidMelaFile
	}

	var recipes []*Recipe
	err = ParseRecipes(f, fs.Size(), func(r *Recipe, inErr error) {
		if inErr == nil {
			recipes = append(recipes, r)
		}
	})

	return recipes, err
}

func withoutExt(name string) string {
	ext := filepath.Ext(name)
	return name[0 : len(name)-len(ext)]
}

// ParseRecipe parses a known single .melarecipe file into a Recipe-compatible struct
func ParseRecipe(r io.Reader) (*Recipe, error) {
	var recipe Recipe

	dec := json.NewDecoder(r)
	err := dec.Decode(&recipe)
	return &recipe, err
}

// ParseRecipe parses a known .melarecipes collection file into a stream of Recipe-compatible structs, calling the onRecipe func for each, as it is parsed
func ParseRecipes(r io.ReaderAt, size int64, onRecipe func(*Recipe, error)) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}

	for _, zf := range zr.File {
		rr, err := zf.Open()
		if err != nil {
			onRecipe(nil, err)
		}
		defer rr.Close()

		if recipe, err := ParseRecipe(rr); err != nil {
			onRecipe(nil, err)
		} else {
			recipe.Filename = withoutExt(zf.Name)
			onRecipe(recipe, nil)
		}
	}

	return nil
}

func (r *Recipe) Save(dir string) (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("unable to marshal recipe: %w", err)
	}

	destination := filepath.Join(dir, r.Filename+".melarecipe")
	f, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("unable to create recipe file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		return "", fmt.Errorf("unable to write data to recipe file: %w", err)
	}

	return destination, nil
}
