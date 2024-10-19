package crouton

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jphastings/crouton-recipes/ingredients"
	"github.com/jphastings/crouton-recipes/uuid"
)

type Recipe struct {
	Filename        string                      `json:"-"`
	UUID            uuid.UUID                   `json:"uuid"`
	Name            string                      `json:"name"`
	SourceName      string                      `json:"sourceName"`
	Ingredients     []ingredients.IngredientUse `json:"ingredients"`
	Serves          PeopleCount                 `json:"serves,omitempty"`
	Duration        Minutes                     `json:"duration,omitempty"`
	CookingDuration Minutes                     `json:"cookingDuration,omitempty"`
	WebLink         Link                        `json:"webLink,omitempty"`
	Steps           Steps                       `json:"steps"`
	Images          []B64Image                  `json:"images,omitempty"`
	Notes           string                      `json:"notes,omitempty"`
	SenderName      string                      `json:"senderName"`

	standardizationsMade []string
}

type Link string

func (r *Recipe) ensureFilename() {
	if r.Filename != "" {
		return
	}

	r.Filename = stringToFilename(r.Name)
}

func (r *Recipe) Save(dir string) (string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("output directory '%s' does not exist", dir)
	}

	data, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("unable to marshal recipe: %w", err)
	}

	r.ensureFilename()

	destination := filepath.Join(dir, stringToFilename(r.SourceName), r.Filename+".crumb")
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return "", fmt.Errorf("unable to create recipe directory '%s': %w", filepath.Dir(destination), err)
	}

	f, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("unable to create recipe file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		return "", fmt.Errorf("unable to write data to recipe file: %w", err)
	}

	return destination, nil
}
