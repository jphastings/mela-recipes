package formats

import "github.com/jphastings/recipes/internal/llm"

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
	New func(Recipe) (Recipe, error)
	// Turns one or more interchange recipes into this collection format
	// Will be nil if this is not a collection format
	NewCollection func(name string, recipes []Recipe) (RecipeCollection, error)
	// Parses a filesystem object into either a single Recipe *or* a single RecipeCollection.
	Parse func(Bundle, ParseOptions) (Recipe, RecipeCollection, error)
	// Bundle must extract sets of recipe files for this format that *must* be processed together.
	// Eg. cooklang stores images adjacent to the recipe file:
	//   lasagne.cook, lasagne.jpg, shakshouka.cook, random.jpg, ignored.crumb
	//   would result in two bundles, the first with one image the second with none, and two unused files.
	Bundle Bundler
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
