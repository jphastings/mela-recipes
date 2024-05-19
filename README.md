# Mela recipes

An opinionated library for stream-parsing [Mela](https://mela.recipes)'s recipe files.

Includes customisations that define a convention for the ID of recipes derived from books. See [ISBN extension](#isbn-extension) for examples.

## Usage

### As a CLI tool

The pre-compiled binaries are [available on Github](https://github.com/jphastings/mela-recipes/releases/latest). You can also install rapidly with Homebrew on Linux and macOS:

```bash
brew install jphastings/tools/mela-standardize
``````

Then standardizing a mela recipe file is as simple as:

```bash
$ mela-standardize recipe1.melarecipe lots.melarecipes /output/path
Saved 'Some recipe' to '/output/path/some-book/some-recipe.melarecipe'
Saved 'A title' to '/output/path/example.com/a-title.melarecipe'
```

### As a library

[![Go Reference](https://pkg.go.dev/badge/github.com/jphastings/mela-recipes.svg)](https://pkg.go.dev/github.com/jphastings/mela-recipes)

```go global
// import github.com/jphastings/mela-recipes
```

The simple `Open` function is quickest for interacting with `.melarecipe` and `.melarecipes` files:

```go ExampleOpen
recipes, err := mela.Open("fixtures/a+b.melarecipes")
if err != nil {
  log.Fatalf("A filesystem error: %v\n", err)
}

for i, r := range recipes {
  fmt.Printf("Recipe #%d title: %s\n", i, r.Title)
}

// Output:
// Recipe #0 title: B title
// Recipe #1 title: A title
```

_Note: the order of the recipes is defined on the structure of the underlying zip file, which isn't necessarily alphabetical, or the sort order of the recipes when exported._

ISBNs can be set & parsed with the `SetBook` and `Book` methods:

```go ExampleSetBook
recipes, err := mela.Open("fixtures/a.melarecipe")
if err != nil {
  log.Fatalf("A filesystem error: %v\n", err)
}

r := recipes[0]

// Note: Setting the book details creates a new object with a URN based on a standardised form ISBN-13.
setErr := r.SetBook("123456789X", mela.MustParsePages("42"), 2)
if setErr != nil {
  log.Fatalf("Invalid Book details given: %v\n", err)
}

fmt.Println("ID:", r.ID)
fmt.Println("ISBN:", r.Book().ISBN13)
fmt.Println("Page numbers:", r.Book().Pages)
fmt.Println("Recipe number:", r.Book().RecipeNumber)


// Output:
// ID: urn:isbn:9781234567897#pages=42&recipe=2
// ISBN: 9781234567897
// Page numbers: 42
// Recipe number: 2
```

You can standardize the Recipe file with a call to `Standardize()`. This performs three standardizations:

- Pulls an ISBN, page & recipe numbers from the _Notes_ field, if present in the form `9781234512345, p.123-125, 2nd` to represent the book with ISBN 9781234512345, optionally on pages 123 to 125, optionally the 2nd recipe on that first page (see [ISBN Extension](#isbn-extension) for more). Changes the recipe's ID to reference this book.
- Converts any images to be maximum 1024x1024px, and in WebP format.
- For books with an ISBN, retrieves the book title from the [OpenLibrary](https://openlibrary.com) and sets the 'link' field of the recipe to be the title of the book.

## Extensions

This library includes backwards-compatible extensions to the [Mela file format](https://mela.recipes/fileformat/index.html).

### ISBN Extension

For recipes that have been scanned or imported from books, the `id` field of the recipe can be set to an ISBN URN with optional page and recipe-number-on-page references. This is invisible to users of `.melarecipe`/`.melarecipes` files, but provides useful information for cataloguing.

For example, the second recipe on page 42 of the book with ISBN-13 `9781234567897` (which would be ISBN-10 `123456789X`) would have an ID of `urn:isbn:9781234567897#pages=42&recipe=1`.

Any `.melarecipe` that has an `id` which is a URN meeting the [RFC-3187 spec](https://www.rfc-editor.org/rfc/rfc3187.txt) will be interpreted as having come from a book.

If that URN includes a valid `pages` f-component (see [RFC-8141§2.3](https://www.ietf.org/rfc/rfc8141.html#section-2.3.3)), then the recipe will be interpreted as being imported from from the page or pages labelled with the specific page numbers.

If the URN _also_ includes a valid `recipe` f-component, then the recipe will be interpreted as coming from the Nth recipe on the first page referenced in `pages`. `0` represents "not explicitly specified, presumed the first recipe", `1` explicitly declares this recipe as the first one on the page, `2` explicitly as the second and so on. (Neatly resolving the awkward difference between humans and machines on zero-indexing).

### Examples

- The first recipe on a single page: `#pages=42` or, explicitly, `#pages=42&recipe=1`
- The second recipe on a single page: `#pages=42&recipe=2`
- (The first recipe on) a range of contiguous pages: `#pages=42-45`
- (The first recipe on) a set of non-contiguous pages: `#pages=42,44,46-49`
- (The first recipe on) a set of pages that use non-numeric numbering: `#pages=v-vii,x-xii`
- (The first recipe on) a page with a number that uses hyphens: `#pages=3%2D2`

The pages referenced should be listed in the order they appear in the book. For example, `#pages=42-41` and `#pages=42,41` would both be incorrect unless the page labelled "41" comes immediately _after_ the page labelled "42" in the direction the book is read).

<details>
  <summary>ABNF notation</summary>

  ```abnf
  pages_f     = contig *( "," contig )
  contig      = page-num [ "-" page-num ]
  page-num    = 1*( ALPHA / DIGIT / pct-encoded )
  pct-encoded = "%" HEXDIG HEXDIG

  recipe_f    = 1*DIGIT
  ```

  (Using [RFC5234 syntax](https://www.rfc-editor.org/rfc/rfc5234.txt).)
</details>
