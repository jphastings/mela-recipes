# Design: schema.org `Recipe` importer (JSON-LD, with microdata / h-recipe fallback)

**Status:** proposed · **Scope:** read (parse) first; write optional later · **Priority:** 1 (highest leverage)

## Summary

Import recipes from saved web pages and any app that emits schema.org `Recipe` data. This
single format absorbs most of the web (nearly every recipe site embeds JSON-LD for Google
rich-results) plus the on-disk format of **Nextcloud Cookbook** (`recipe.json`) and exports
from **Tandoor** / **RecipeSage**. It is the highest-value addition because one importer
covers a huge fraction of the ecosystem.

Implement as a single package `schemaorg/` (or `webrecipe/`). Primary path: parse JSON-LD
embedded in HTML. Fallback within the same package, when no JSON-LD `Recipe` is present:
microdata (`itemtype=".../Recipe"`, `itemprop=...`) then the h-recipe microformat
(`class="h-recipe"`, `p-ingredient`, `e-instructions`, `p-yield`, `dt-duration`).

This format is read-only at first (like [epub/](../../epub)): omit `Import`/`NewCollection`,
set only the parse `Features`. JSON-LD *write* is cheap and can be added later.

## References

- schema.org Recipe type: <https://schema.org/Recipe>
- Google recipe structured-data guide: <https://developers.google.com/search/docs/appearance/structured-data/recipe>
- h-recipe microformat: <https://microformats.org/wiki/h-recipe>
- Nextcloud Cookbook stores each recipe as a schema.org `recipe.json` in its own folder
  alongside image files.

## Format details

- **Extensions to bundle:** `.html`, `.htm`, and `.json` (raw JSON-LD / Nextcloud
  `recipe.json`). Use `formats.BundleByExtension(".html", ".htm", ".json")`.
- **Encoding:** JSON. In HTML it lives in one or more
  `<script type="application/ld+json">` blocks. Parse every such block; the document may
  also wrap nodes in `{"@graph": [...]}` or be a top-level array. Walk all nodes and select
  the one whose `@type` is (or contains) `"Recipe"` — `@type` may be a string **or** an
  array of strings.
- **`recipeInstructions` is the awkward field.** It can be any of:
  - a single string (plain text or HTML) → split into steps on newlines / `<p>`/`<li>`,
  - an array of strings,
  - an array of `HowToStep` objects (`{ "@type":"HowToStep", "text":"..." }`),
  - an array of `HowToSection` objects (`{ "@type":"HowToSection", "name":"...",
    "itemListElement":[HowToStep...] }`) — these map directly onto sectioned
    `[]TitledList`.
- **Durations are ISO 8601** (e.g. `"PT1H30M"`). `formats.MaybeDuration` does **not** parse
  these (it handles `"30 min"`); use an ISO-8601 duration parser (e.g.
  `github.com/sosodev/duration`, or a tiny hand-rolled `PnHnMnS` parser).
- **`image`** can be a string URL, an array of URLs, or an `ImageObject` (`{ "url": ... }`).

## Mapping to `InterchangeRecipe`

[internal/formats/recipe.go](../../internal/formats/recipe.go) defines the target.

| schema.org property | Interchange field | Notes |
|---|---|---|
| `name` | `Title` | required by `Validate()` |
| `description` | `Description` | |
| `recipeYield` | `Yield` | may be number/string/array → stringify first element |
| `recipeIngredient` (or legacy `ingredients`) | `Ingredients` | one `TitledList{Title:"", List: [...]}` (flat strings) |
| `recipeInstructions` | `Instructions` | strings/`HowToStep` → one section; `HowToSection` → one `TitledList` per section using `name` as the title |
| `prepTime` / `cookTime` / `totalTime` | `PrepTime` / `CookTime` / `TotalTime` | ISO-8601 → `*time.Duration` |
| `image` | `Images` | URL(s) — see decision below |
| `recipeCategory`, `recipeCuisine`, `keywords` | `Tags` | merge; `keywords` may be a comma-separated string |
| `author`, `datePublished`, `aggregateRating`, `video` | — | no interchange field; drop |
| `nutrition` | `Notes` (append) or drop | interchange has no nutrition field |

## Implementation notes (follow the [mela/](../../mela) package shape)

- `recipe.go`: a `Recipe` struct holding the already-mapped `InterchangeRecipe` (or the raw
  parsed JSON-LD). Implement `formats.Recipe`:
  - `Name`, `Format`, `Filename` (use `standardize.StringToFilename(title)` +
    `FormatInfo.Extension` like Mela does);
  - `Export()` → return the mapped `InterchangeRecipe`;
  - `Marshal(io.Writer)` → return a "not yet implemented" error initially (read-only), or
    emit JSON-LD when write is added;
  - `Standardize()` → mirror Mela's standardize (filename-from-title, image optimise); can
    start as a no-op returning `nil, nil`.
- `parse.go`: a `Parse` [`formats.Parser`](../../internal/formats/types.go) that reads the
  bundle's file, extracts the `Recipe` node, maps it, and emits a single
  `ParseEvent{Recipe, Err, I:1, N:1}` on a goroutine-fed channel (copy the structure of
  [mela/parse.go](../../mela/parse.go)).
- `format.go`: `FormatInfo = &formats.Format{ Name: "schema.org", URL: "https://schema.org/Recipe",
  Features: formats.Features{ParseRecipe: true}, Extension: ".html", Parse: Parse,
  Bundle: formats.BundleByExtension(".html", ".htm", ".json") }`. Register it in
  [`AvailableFormats()`](../../formats.go).
- Suggested deps: `golang.org/x/net/html` (DOM walk to find script blocks / microdata),
  `encoding/json`. A microformats parser (e.g. `willnorris.com/go/microformats`) is optional
  and only needed for the h-recipe fallback.

## Edge cases & decisions

- **Images are URLs, not bytes.** `Images []utils.B64Image` expects image data. Decide:
  (a) store nothing and skip images, (b) fetch the URL and wrap as `B64Image` so the
  existing image-optimise standardisation applies. Fetching is a network op — gate it behind
  a `ParseOptions` flag, defaulting to skip. (`ParseOptions` already carries an `*llm.Connection`;
  a `FetchImages bool` could be added there.)
- **Pick one Recipe per document.** If multiple `Recipe` nodes exist, take the first valid
  one (or emit each as a separate `ParseEvent`).
- **HTML entities & embedded markup** in instructions/ingredients should be unescaped and
  tag-stripped to plain text.
- Run `InterchangeRecipe.Validate()` and surface a clear error when title / ingredients /
  instructions are missing rather than emitting an empty recipe.

## Test fixtures

Add `schemaorg/fixtures/` with: a real-world page containing JSON-LD; a page with only
microdata; a page with only h-recipe; a Nextcloud `recipe.json`; and one with a
`HowToSection`-structured `recipeInstructions`. Table-test `Parse` → `Export` against the
expected `InterchangeRecipe` (title, section counts, durations, tags).
