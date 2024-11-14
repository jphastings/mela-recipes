package formats

import (
	"os"
	"time"
)

// Details about a recipe format
type Format struct {
	// A human friendly (eg. capitalised) name for the format
	Name string
	// An official URL pointing to the product that uses this format or the official documentation for the format — whichever will be the most informative for users trying to learn about this format.
	URL string
	// What features this format provider has
	Features Features
	// The file extension for the recipe format (without period)
	Extension string
	// The file extension for the collection format (without period)
	ExtensionCollection string
	// Turns one interchange recipe format into this format (if it exists, without period)
	New func(InterchangeRecipe) (Recipe, error)
	// Turns one or more interchange recipes into this collection format
	// Will be nil if this is not a collection format
	NewCollection func(string, []InterchangeRecipe) (RecipeCollection, error)
	// Parses a filesystem object into either a single Recipe *or* a single RecipeCollection.
	// A response of nil, nil, nil means the file is definitively not of this format.
	Parse func(*os.File) (Recipe, RecipeCollection, error)
}

type Features struct {
	ParseRecipe     bool
	WriteRecipe     bool
	ParseCollection bool
	WriteCollection bool
}

// A generic and internal structure for recipes that is used for conversion
// ⚠️ This struct is highly likely to change subtly with each new recipe format added to this library.
type InterchangeRecipe struct {
	Filename    string
	ID          string
	Title       string
	Description string

	PrepTime  time.Duration
	CookTime  time.Duration
	TotalTime time.Duration
}
