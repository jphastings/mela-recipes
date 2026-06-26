# Design: RecipeML (`.xml`) importer

**Status:** proposed · **Scope:** read-only import · **Priority:** 3 (legacy archives)

## Summary

RecipeML (formerly "DESSERT") is an open, published XML recipe format from the early 2000s.
The ecosystem is effectively dead, but the spec is clean and well-structured (quantity / unit
/ item are separate nodes) and there are sizeable archives of `.xml` recipe files. Implement
as a **read-only** package `recipeml/`: omit `Import`/`NewCollection`, set only `ParseRecipe`,
and have `Marshal` return a "not supported" error. A document may contain multiple
`<recipe>` elements → emit one `ParseEvent` each.

## References

- Specification (DTD/XSD): <http://www.formatdata.com/recipeml/spec/recipeml-spec.html>
- Reference parsers to crib from: <https://github.com/samjavner/recipeformats>

## Format details

```xml
<recipeml version="0.5">
  <recipe>
    <head>
      <title>Chocolate Chip Cookies</title>
      <categories><cat>Dessert</cat><cat>Cookies</cat></categories>
      <yield><range><q>24</q></range></yield>
    </head>
    <ingredients>
      <ing-div title="Dough">
        <ing><amt><qty>2</qty><unit>cup</unit></amt><item>flour</item></ing>
        <ing><amt><qty>1/2</qty><unit>tsp</unit></amt><item>salt</item><prep>sifted</prep></ing>
      </ing-div>
    </ingredients>
    <directions>
      <step>Cream the butter and sugar.</step>
      <step>Bake at 375F for 10 minutes.</step>
    </directions>
    <description>Classic cookies.</description>
    <note>Best fresh.</note>
  </recipe>
</recipeml>
```

- Root `<recipeml>` may hold one or many `<recipe>` elements.
- `<ing-div>` groups ingredients into optionally-titled sections.
- Each `<ing>` separates `<qty>`, `<unit>`, `<item>`, and optional `<prep>` / `<modifier>`.

## Mapping to `InterchangeRecipe`

[internal/formats/recipe.go](../../internal/formats/recipe.go)

| RecipeML | Interchange field | Notes |
|---|---|---|
| `head/title` | `Title` | required |
| `description` | `Description` | |
| `head/categories/cat` | `Tags` | one per `<cat>` |
| `head/yield` | `Yield` | reconstruct from `<range>/<q>` (+ any unit) |
| `ingredients/ing-div` + `ing` | `Ingredients` | one `TitledList` per `<ing-div>` (use its `title` attr); reconstruct each item as `"qty unit item, prep"` |
| `directions/step` | `Instructions` | one `TitledList{Title:"", List: step texts}` |
| `note` | `Notes` | |
| — | `Images` | none typically |

## Implementation notes (follow [mela/](../../mela), but read-only)

- `recipe.go`: `Recipe` struct + `formats.Recipe` methods; `Marshal` returns a "not
  supported" error; `Standardize` can be a no-op.
- `parse.go`: `Parse` [`formats.Parser`](../../internal/formats/types.go) using
  `encoding/xml` with structs mirroring the elements above; iterate `<recipe>` nodes and emit
  a `ParseEvent` per recipe (`I`/`N`). `CollectionDetails` is `nil`.
- `format.go`: `FormatInfo = &formats.Format{ Name: "RecipeML",
  URL: "http://www.formatdata.com/recipeml/", Features: formats.Features{ParseRecipe: true},
  Extension: ".xml", Parse: Parse, Bundle: formats.BundleByExtension(".xml") }`. Register in
  [`AvailableFormats()`](../../formats.go).

## Edge cases & decisions

- **`.xml` is generic.** Bundling all `.xml` files will catch non-recipe XML. Have `Parse`
  confirm the root element is `<recipeml>` (or the RecipeML DOCTYPE) and emit a clear
  non-fatal error otherwise. Because the [`Bundler`](../../internal/formats/bundle.go) runs
  before parsing and consumes files from later formats, decide whether `.xml` should be
  claimed eagerly or only sniffed — document the choice.
- **DOCTYPE / external DTD:** disable external entity resolution (`xml.Decoder` does not fetch
  DTDs by default — keep it that way to avoid XXE).
- **Quantity formatting:** `<qty>` values like `1/2` are plain text — keep as-is in the
  reconstructed item string.
- Run `InterchangeRecipe.Validate()`; report missing title/ingredients/instructions.

## Test fixtures

Add `recipeml/fixtures/` with a single-recipe `.xml`, a multi-recipe document, and one using
multiple `<ing-div>` sections. Assert recipe count, section titles, and reconstructed
ingredient strings.
