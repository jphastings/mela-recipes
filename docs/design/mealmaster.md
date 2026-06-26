# Design: MealMaster (`.mmf`) importer

**Status:** proposed · **Scope:** read-only import · **Priority:** 3 (legacy archives)

## Summary

MealMaster is a 1990s DOS recipe program whose plain-text export (`.mmf`) holds an enormous
body of BBS-era recipe archives (hundreds of thousands of recipes). There is no formal spec —
the format is reverse-engineered — and the fixed-width, sometimes two-column ingredient
layout makes it the most fragile of the parsers here. Implement as a **read-only** package
`mealmaster/`: omit `Import`/`NewCollection`, set only `ParseRecipe`, and have `Marshal`
return a "not supported" error.

A single `.mmf` file commonly concatenates **many** recipes, so `Parse` emits one
`ParseEvent` per recipe with a running `I` and total `N`.

## References

- Reverse-engineered grammar & walkthrough: <https://www.wedesoft.de/software/2020/07/07/mealmaster/>
- Canonical format notes: <https://www.ffts.com/mmformat.txt>
- Existing parsers to crib field handling from: <https://github.com/samjavner/recipeformats>

## Format details

```
---------- Recipe via Meal-Master (tm) v8.05
 
      Title: Chocolate Chip Cookies
 Categories: Dessert, Cookies
      Yield: 24 cookies
 
      2 c  Flour
    1/2 ts Salt
      1 c  Butter, softened           ← amount(cols 1-7) unit(cols 9-10) text(cols 12+)
      2 ea Eggs
 
  Cream butter and sugar.  Add eggs...   ← free-text directions
  Bake at 375F for 10 minutes.
 
-----                                     ← recipe terminator (or `MMMMM`)
```

- **Header line:** begins with ≥5 hyphens and contains `Meal-Master`. Marks the start of a
  recipe.
- **Labelled fields:** `Title:`, `Categories:` (comma list), `Yield:`.
- **Ingredients block:** fixed columns — amount (chars 1–7), unit (chars 9–10, a 2-letter
  code), ingredient text (chars 12+). Continuation lines start with `-` in the text column
  and append to the previous ingredient. Some files lay ingredients out in **two side-by-side
  columns** (~40 chars each).
- **Section headers** inside ingredients appear as a centred line wrapped in `-----` or as
  `MMMMM-----Heading-----`.
- **Directions:** free text after the ingredients block.
- **Terminator:** a line of hyphens (`-----`) or `MMMMM`. Variants use `MMMMM` as both the
  header and footer marker.

## Mapping to `InterchangeRecipe`

[internal/formats/recipe.go](../../internal/formats/recipe.go)

| MealMaster | Interchange field | Notes |
|---|---|---|
| `Title:` | `Title` | required |
| `Categories:` | `Tags` | split on commas |
| `Yield:` | `Yield` | |
| ingredient lines | `Ingredients` | reconstruct `"amount unit text"` per line; section headers → separate `TitledList`s, otherwise one `Title:""` section |
| directions text | `Instructions` | one `TitledList{Title:"", List: paragraphs}` |
| — | `Images` | none in this format |

## Implementation notes (follow [mela/](../../mela), but read-only)

- `recipe.go`: `Recipe` struct + `formats.Recipe` methods; `Marshal` returns
  `fmt.Errorf("marshalling MealMaster not supported")`, `Standardize` can be a no-op.
- `parse.go`: `Parse` [`formats.Parser`](../../internal/formats/types.go) that scans the file
  splitting on header/terminator markers, parses each block, and emits a `ParseEvent` per
  recipe (`I` increments, `N` = total recipes found — count them up-front or stream with `N`
  growing). Return `CollectionDetails` as `nil` (a `.mmf` is a flat archive, not a named
  app collection).
- `format.go`: `FormatInfo = &formats.Format{ Name: "MealMaster",
  URL: "https://www.ffts.com/mmformat.txt", Features: formats.Features{ParseRecipe: true},
  Extension: ".mmf", Parse: Parse, Bundle: formats.BundleByExtension(".mmf") }`. Register in
  [`AvailableFormats()`](../../formats.go).

## Edge cases & decisions

- **Two-column ingredient layout** is the main hazard — detect it (two amount/unit blocks on
  one line) and split, or accept single-column only in a first pass and log unsupported
  layouts.
- **Continuation lines** (`-` prefix) must be folded into the prior ingredient.
- **Character encoding:** these files are often CP437 / Latin-1, not UTF-8 — decode
  accordingly (fractions like `½`, degree signs).
- **Unit codes** are 2-letter abbreviations (`ts`, `tb`, `c`, `ea`, `pk`...). A lookup table
  expanding them to full words is optional; keeping the raw code is acceptable for v1.
- Be liberal in what you accept: malformed blocks should produce a `ParseEvent` with an
  `Err` rather than aborting the whole file.

## Test fixtures

Add `mealmaster/fixtures/` with a single-recipe `.mmf`, a multi-recipe `.mmf`, and one using
the two-column ingredient layout. Assert recipe count, titles, ingredient counts, and that a
continuation line folds correctly.
