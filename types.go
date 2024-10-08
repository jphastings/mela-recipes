package recipes

import "time"

// Details about a recipe format
type Format struct {
	// A human friendly (eg. capitalised) name for the format
	Name string
	// The file extension for the format (without period)
	Extension string
	// Whether this format is a collection format
	IsCollection bool
	// Turns an interchange recipe format into this format
	Import func(InterchangeRecipe) (Recipe, error)
}

// A generic and internal structure for recipes that is used for conversion
// ⚠️ This struct is likely to change subtly with each new recipe format added to this library
type InterchangeRecipe struct {
	Filename    string
	ID          string
	Title       string
	Description string

	PrepTime  time.Duration
	CookTime  time.Duration
	TotalTime time.Duration
}
