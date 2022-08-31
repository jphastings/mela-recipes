# Mela recipes

An opinionated library for stream-parsing [Mela](https://mela.recipes)'s recipe files.

Includes customisations that define a convention for the ID of recipes derrived from books. For example, a recipe found on page 42 of the book with ISBN-13 `9781234567897` (which would be ISBN-10 `123456789X`) would have an ID of `urn:isbn:9781234567897#page=42`.

## Usage

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
  fmt.Printf("Recipe #%d title: %s\n", i + 1, r.Title())
}

// Output:
// Recipe #1 title: B title
// Recipe #2 title: A title
```

_Note: the order of the recipes is defined on the structure of the underlying zip file, which isn't necessarily alphabetical, or the sort order of the recipes when exported._

ISBNs can be set & parsed with the `SetBook` and `Book` methods:

```go ExampleSetBook
recipes, err := mela.Open("fixtures/a.melarecipe")
if err != nil {
  log.Fatalf("A filesystem error: %v\n", err)
}

// Note: Setting the book details creates a new object with a URN based on a standardised form ISBN-13.
r, err := recipes[0].SetBook("123456789X", 42)
if err != nil {
  log.Fatalf("Invalid ISBN given: %v\n", err)
}

fmt.Println("ID:", r.ID())
fmt.Println("ISBN:", r.Book().ISBN13)
fmt.Println("Page number:", r.Book().Page)


// Output:
// ID: urn:isbn:9781234567897#page=42
// ISBN: 9781234567897
// Page number: 42
```
