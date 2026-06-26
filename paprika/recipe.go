package paprika

import (
	"compress/gzip"
	"encoding/json"
	"io"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/utils"
)

var _ formats.Recipe = (*Recipe)(nil)

// Recipe is a Paprika Recipe Manager recipe. On disk it is gzip-compressed JSON
// (a .paprikarecipe file, or one entry within a .paprikarecipes zip). Fields the
// interchange format can't represent (nutrition, source, rating, …) are kept so
// they survive Paprika→Paprika round-trips.
type Recipe struct {
	filename string `json:"-"`

	UID             string                    `json:"uid"`
	Title           string                    `json:"name"`
	Description     string                    `json:"description,omitempty"`
	Ingredients     formats.SectionedSequence `json:"ingredients"`
	Directions      formats.SectionedSequence `json:"directions"`
	Notes           string                    `json:"notes,omitempty"`
	NutritionalInfo string                    `json:"nutritional_info,omitempty"`
	Servings        string                    `json:"servings,omitempty"`
	Source          string                    `json:"source,omitempty"`
	SourceURL       string                    `json:"source_url,omitempty"`
	Categories      []string                  `json:"categories"`
	PhotoData       utils.B64Image            `json:"photo_data,omitempty"`
	ImageURL        string                    `json:"image_url,omitempty"`
	PrepTime        formats.MaybeDuration     `json:"prep_time,omitempty"`
	CookTime        formats.MaybeDuration     `json:"cook_time,omitempty"`
	TotalTime       formats.MaybeDuration     `json:"total_time,omitempty"`
	Rating          int                       `json:"rating,omitempty"`
	Difficulty      string                    `json:"difficulty,omitempty"`
	Created         string                    `json:"created,omitempty"`
	Hash            string                    `json:"hash,omitempty"`
	OnFavorites     bool                      `json:"on_favorites,omitempty"`
}

func (r Recipe) Name() string            { return r.Title }
func (r Recipe) Format() *formats.Format { return FormatInfo }
func (r Recipe) Filename() string {
	if r.filename == "" {
		return standardize.StringToFilename(r.Title) + FormatInfo.Extension
	}
	return r.filename + FormatInfo.Extension
}

// ParseRecipeStream decodes a single recipe from gzip-compressed Paprika JSON.
func ParseRecipeStream(r io.Reader) (*Recipe, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return parseRecipeJSON(gz)
}

func parseRecipeJSON(r io.Reader) (*Recipe, error) {
	var recipe Recipe
	if err := json.NewDecoder(r).Decode(&recipe); err != nil {
		return nil, err
	}
	return &recipe, nil
}

// Marshal writes the recipe as gzip-compressed JSON — the encoding Paprika uses
// for both standalone .paprikarecipe files and entries inside a .paprikarecipes
// archive.
func (r *Recipe) Marshal(w io.Writer) error {
	gz := gzip.NewWriter(w)
	if err := json.NewEncoder(gz).Encode(r); err != nil {
		gz.Close()
		return err
	}
	return gz.Close()
}
