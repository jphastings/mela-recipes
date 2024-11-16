package crouton

import (
	"encoding/json"
	"io"

	"github.com/jphastings/recipes/crouton/ingredients"
	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/internal/uuid"
)

var _ formats.Recipe = (*Recipe)(nil)

type Recipe struct {
	filename        string                      `json:"-"`
	UUID            uuid.UUID                   `json:"uuid"`
	RecipeName      string                      `json:"name"`
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
}

func (r Recipe) Name() string            { return r.RecipeName }
func (r Recipe) Format() *formats.Format { return FormatInfo }
func (r Recipe) Filename() string        { return r.filename + FormatInfo.Extension }

type Link string

func (r *Recipe) ensureFilename() {
	if r.filename != "" {
		return
	}

	r.filename = standardize.StringToFilename(r.RecipeName)
}

func (r *Recipe) Marshal(w io.Writer) error {
	return json.NewEncoder(w).Encode(r)
}
