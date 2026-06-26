# Design: RecipeMD importer/exporter

**Status:** proposed · **Scope:** read **and** write · **Priority:** 2

## Summary

RecipeMD is an open, actively-maintained Markdown recipe standard (CommonMark-based) with a
reference implementation and CLI. It is human-writable and git-friendly, which makes it both
a good import source and an ideal round-trip / storage **write target**. Implement as a
package `recipemd/` following the [mela/](../../mela) shape, supporting both `Parse` and
`Import` + `Marshal`.

## References

- Specification: <https://recipemd.org/specification.html>
- Reference implementation & samples: <https://github.com/RecipeMD/RecipeMD>

## Format details

Extension `.md`. A RecipeMD document has a fixed structure:

```markdown
# Recipe Title

A short description paragraph (optional, one or more paragraphs).

*tag one, tag two*          ← tags: a paragraph that is only an italic comma-list
**4 servings, 1 cake**      ← yield: a paragraph that is only a bold comma-list

---                         ← first horizontal rule

- *2 g* salt                ← ingredients; amount in italics, rest is the item
- *1* onion, diced

## Sauce                     ← optional ingredient-group headers → sections
- *200 ml* cream

---                         ← second horizontal rule

Free-form Markdown instructions go here, paragraphs and/or numbered steps.
```

- **Title:** the first level-1 header (`# ...`).
- **Description:** paragraph(s) between the title and the first `---`.
- **Tags:** a paragraph containing only a comma-separated list in *italics*.
- **Yield:** a paragraph containing only a comma-separated list in **bold**.
- **Ingredients:** list items between the first and second `---`; optional `##`/`###`
  headers group them into sections. Each item encodes the amount in italics
  (`*2 g* salt`).
- **Instructions:** everything after the second `---` (arbitrary Markdown).

## Mapping to `InterchangeRecipe`

[internal/formats/recipe.go](../../internal/formats/recipe.go)

| RecipeMD element | Interchange field | Notes |
|---|---|---|
| `# Title` | `Title` | required |
| description paragraphs | `Description` | |
| italic comma-list | `Tags` | |
| bold comma-list | `Yield` | join list back to a string |
| ingredient lists (+ `##` group headers) | `Ingredients` | one `TitledList` per group; ungrouped items go in a `Title:""` section. Keep each item as a single string (`"2 g salt"`) — interchange has no structured amount/unit |
| instructions Markdown | `Instructions` | one `TitledList{Title:"", List: steps}`; split on numbered list items / paragraphs, or on `##` subheaders into multiple sections |
| inline `![](img)` | `Images` | optional; usually omitted |

Writing (`Import` + `Marshal`) reverses this: emit `# Title`, description, `*tags*`,
`**yield**`, `---`, ingredient lists (re-italicising the amount if present), `---`, then
instructions. Round-trip should be stable for recipes authored in RecipeMD.

## Implementation notes (follow [mela/](../../mela))

- `recipe.go`: `Recipe` struct implementing [`formats.Recipe`](../../internal/formats/types.go) —
  `Name`, `Format`, `Filename`, `Export()`, `Marshal()` (writes the Markdown), `Standardize()`.
- `parse.go`: `Parse` [`formats.Parser`](../../internal/formats/types.go) emitting one
  `ParseEvent{Recipe, Err, I:1, N:1}` (cf. [mela/parse.go](../../mela/parse.go)).
- `import.go`: `Import func(formats.Recipe) (formats.Recipe, error)` — `Export()` the source
  to interchange, build a RecipeMD `Recipe` from it.
- `format.go`: `FormatInfo = &formats.Format{ Name: "RecipeMD",
  URL: "https://recipemd.org", Features: formats.Features{ParseRecipe: true, WriteRecipe: true},
  Extension: ".md", Import: Import, Parse: Parse, Bundle: ... }`. Register in
  [`AvailableFormats()`](../../formats.go).
- Suggested dep: `github.com/yuin/goldmark` (CommonMark parser) to get an AST, then walk it
  for the structural rules above rather than regexing raw Markdown.

## Edge cases & decisions

- **`.md` is generic.** Bundling every `.md` file risks grabbing unrelated Markdown. Either
  bundle `.md` and have `Parse` reject documents that don't match the RecipeMD shape (no H1,
  missing the two `---` separators, no ingredient list) with a clear non-fatal error, or
  require an explicit opt-in. Document the chosen behaviour.
- Tags/yield detection must check a paragraph is *entirely* italic / bold comma-list, not
  just that it contains emphasis.
- Amounts: keep the italic amount inline in the item string for now; structured parsing of
  amount/unit is out of scope (interchange can't hold it anyway).
- Run `InterchangeRecipe.Validate()` and report missing title/ingredients/instructions.

## Test fixtures

Add `recipemd/fixtures/` with a couple of real RecipeMD files (including one using
ingredient-group headers). Test `Parse` → `Export` field mapping, and a `Import` → `Marshal`
→ `Parse` round-trip that returns to the same `InterchangeRecipe`.
