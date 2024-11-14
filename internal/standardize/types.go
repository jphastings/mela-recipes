package standardize

// A type of a standardisation that can be made to a recipe
type Std string

var (
	// Extracts & standardises the ISBN and other physical book data found in the recipe
	StdISBN Std = "isbn"
	// Resizes and optimises images associated with the recipe
	StdImages Std = "images"
	// Derives the filename from the recipe title
	StdFilename Std = "filename"
)
