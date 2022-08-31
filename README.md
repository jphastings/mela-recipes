# Mela recipes

An opinionated library for stream-parsing [Mela](https://mela.recipes)'s recipe files.

```go ExampleOpen
// import github.com/jphastings/mela-recipes

onRecipe := func(r mela.Recipe, err error) {
  if err != nil {
    fmt.Printf("An invalid recipe: %v\n", err)
  }

  fmt.Printf("Recipe title: %s\n", r.Title())
}

fsErr := mela.Open("fixtures/a+b.melarecipes", onRecipe)
if fsErr != nil {
  fmt.Printf("A filesystem error: %v\n", fsErr)
}

// Output:
// Recipe title: B title
// Recipe title: A title
```
