package recipes

// Represents a recipe in a specific format
type Recipe interface {
	// The format this recipe is in
	Format() Format
	// Returns the stored filename (with extension), or an appropriate generated one
	Filename() string
	// Destructively applies all standardizations to this recipe.
	// Typically this includes:
	// - Setting the filename from the recipe title
	// - Optimising any images attached
	// - Extracting and standardising the ISBN & physical book data found in the notes field, if one is present
	Standardize() error
	// Mutates the recipe in this format into the interchange format
	ToInterchange() (InterchangeRecipe, error)
	// Creates a recipe in this format from the provided interchange recipe
	FromInterchange(InterchangeRecipe) (Recipe, error)
}

// Represents a collection of recipes held in a specific format
type RecipeCollection interface {
	// The format this recipe collection is in
	Format() Format
	// Returns the stored filename (with extension), or an appropriate generated one
	Filename() string
	// Adds the provided recipe to this collection
	Add(...*Recipe) error
	// Declares that no more recipes will be added
	Close() error
}
