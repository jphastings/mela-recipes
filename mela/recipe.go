package mela

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"net/url"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/utils"
)

var _ formats.Recipe = (*Recipe)(nil)

type Recipe struct {
	filename     string            `json:"-"`
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Link         string            `json:"link"`
	Text         string            `json:"text"`
	Ingredients  SectionedSequence `json:"ingredients"`
	Instructions SectionedSequence `json:"instructions"`
	Nutrition    string            `json:"nutrition"`
	Categories   []string          `json:"categories"`
	Notes        string            `json:"notes"`

	Images    []utils.B64Image `json:"images"`
	Yield     PeopleCount      `json:"yield"`
	PrepTime  MaybeDuration    `json:"prepTime"`
	CookTime  MaybeDuration    `json:"cookTime"`
	TotalTime MaybeDuration    `json:"totalTime"`
}

func (r Recipe) Name() string           { return r.Title }
func (r Recipe) Format() formats.Format { return format }
func (r Recipe) Filename() string       { return r.filename + "." + format.Extension }

var ErrInvalidMelaFile = errors.New("given file is neither a melarecipe nor a melarecipes file")
var ErrInvalidMelaRecipeFile = errors.New("given file is not a melarecipe file")
var ErrInvalidMelaRecipesFile = errors.New("given file is not a melarecipes file")

//go:embed melarecipe.schema.json
var JSONSchema string

// ParseRecipe parses a known single .melarecipe file into a Recipe-compatible struct
func ParseRecipe(r io.Reader) (*Recipe, error) {
	var recipe Recipe

	dec := json.NewDecoder(r)
	err := dec.Decode(&recipe)
	return &recipe, err
}

func sourceName(linkField string) string {
	u, err := url.Parse(linkField)
	if err == nil && u.Host != "" {
		return u.Host
	}

	return standardize.StringToFilename(linkField)
}

func (r *Recipe) Marshal(w io.Writer) error {
	return json.NewEncoder(w).Encode(r)
}
