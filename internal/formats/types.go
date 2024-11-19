package formats

import (
	"github.com/jphastings/recipes/internal/llm"
)

// Details about a recipe format
type Format struct {
	// A human friendly (eg. capitalised) name for the format
	Name string
	// An official URL pointing to the product that uses this format or the official documentation for the format — whichever will be the most informative for users trying to learn about this format.
	URL string
	// What features this format provider has
	Features Features
	// The file extension for the recipe format (with period)
	Extension string
	// The file extension for the collection format (with period)
	ExtensionCollection string
	// Turns one interchange recipe format into this format
	Import func(Recipe) (Recipe, error)
	// Creates a new Collection in this format, ready to add recipes into
	NewCollection func(CollectionDetails) (CollectionWriter, error)
	// Parses a filesystem object into either a single Recipe *or* a single RecipeCollection.
	Parse Parser
	// Bundle must extract sets of recipe files for this format that *must* be processed together.
	// Eg. cooklang stores images adjacent to the recipe file:
	//   lasagne.cook, lasagne.jpg, shakshouka.cook, random.jpg, ignored.crumb
	//   would result in two bundles, the first with one image the second with none, and two unused files.
	Bundle Bundler
}

// A Parser function returns information about the collection that is parsed (if it is a collection that was parsed), and a channel that streams recipes as they are decoded
type Parser func(Bundle, ParseOptions) (<-chan ParseEvent, *CollectionDetails, error)

// Every parser will return a channel of ParseEvents. Each of which contains progress information and either a recipe or an error
type ParseEvent struct {
	Recipe Recipe
	Err    error
	// This many have just been parsed
	I int
	// The total to parse has changed to this (zero value will be ignored)
	N int
}

// An indication of the features a format's implementation can provide
type Features struct {
	ParseRecipe     bool
	WriteRecipe     bool
	ParseCollection bool
	WriteCollection bool
}

// The list of the extensions (with periods) the format works with
func (f Format) Extensions() []string {
	var out []string
	if f.Extension != "" {
		out = append(out, f.Extension)
	}
	if f.ExtensionCollection != "" {
		out = append(out, f.ExtensionCollection)
	}
	return out
}

type ParseOptions struct {
	LLM *llm.Connection
}
