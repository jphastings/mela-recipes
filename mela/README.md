# Mela recipes

An opinionated library for stream-parsing [Mela](https://mela.recipes)'s recipe files.

Includes customisations that define a convention for the ID of recipes derived from books. See [ISBN extension](#isbn-extension) for examples.

## Usage

### As a CLI tool

The pre-compiled binaries are [available on Github](https://github.com/jphastings/recipes/releases/latest). You can also install rapidly with Homebrew on Linux and macOS:

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

[![Go Reference](https://pkg.go.dev/badge/github.com/jphastings/recipes.svg)](https://pkg.go.dev/github.com/jphastings/recipes)

```go global
// import github.com/jphastings/recipes/mela
// import github.com/jphastings/recipes/recipecommon
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
setErr := r.SetBook("123456789X", recipecommon.MustParsePages("42"), 2)
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

- Pulls an ISBN, page & recipe numbers from the _Notes_ field, if present in forms similar to `_9781234512345, p.123-125, 2nd_`. This would represent the book with ISBN 9781234512345, on pages 123 to 125, starting as the 2nd recipe on that first page (see [ISBN Extension](#isbn-extension) for more). Changes the recipe's ID to reference this book.
- Converts any images to be maximum 512x512px, and in (jpegli encoded) JPEG format.
- (If network access is enabled, and for books with an ISBN) retrieves the book title from the [OpenLibrary](https://openlibrary.com) and sets the 'link' field of the recipe to be the title of the book.

## Extensions

This library includes extensions to the [Mela file format](https://mela.recipes/fileformat/index.html).

- Supporting [ISBNs](#isbn-extension)
- Supporting [Proof of Ownership](#proof-of-ownership-extension)

### ISBN Extension

For recipes that have been scanned or imported from books, the `id` field of the recipe can be set to an ISBN URN with optional page and recipe-number-on-page references. This is currently invisible to users of `.melarecipe`/`.melarecipes` files, but provides useful information for cataloguing.

For example, the second recipe on page 42 of the book with ISBN-13 `9781234567897` (which would be ISBN-10 `123456789X`) would have an ID of `urn:isbn:9781234567897#pages=42&recipe=1`.

Any `.melarecipe` that has an `id` which is a URN meeting the [RFC-3187 spec](https://www.rfc-editor.org/rfc/rfc3187.txt) will be interpreted as having come from a book.

If that URN includes a valid `pages` f-component (see [RFC-8141§2.3](https://www.ietf.org/rfc/rfc8141.html#section-2.3.3)), then the recipe will be interpreted as being imported from from the page or pages labelled with the specific page numbers.

If the URN _also_ includes a valid `recipe` f-component, then the recipe will be interpreted as coming from the Nth recipe on the first page referenced in `pages`. `0` represents "not explicitly specified, presumed the first/only recipe on this page", `1` explicitly declares this recipe as the first one on the page, `2` explicitly as the second and so on. (Neatly resolving the awkward difference between humans and machines on zero-indexing).

#### ISBN examples

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

### Proof of Ownership Extension

For recipe bundles created from physical books it can be useful to protect the bundle so that only people who are able to demonstrate they own a copy of that physical book are able to access the recipes within.

Files protected like this are identical to `.melarecipes` zip files, except that they use the extension `.protectedrecipes`, all contained `.melarecipe` files are password protected with standard AES256 zip encryption, and also contain the unencrypted `_decrypting.txt` as the first entry, which contains the questions that need to be answered and details needed to derive the decryption password.

The `_decrypting.txt` file contains three sections (each line `\n` delimited):

1. `Q` Questions that can be provided to be answered (one per line, in the main locale of the recipes),
2. An explanatory comment line (starting with `#`) on line number `Q`, and
3. `P` Additional points (one per line, as base64 encoded big endian bigint bytes).

The number of additional points in the last section (`P`) is equal to the number of questions that can remain unanswered while still retrieving a correct decryption key — ie. `Q - P` questions must be answered. To produce a decryption key this process is followed:

- `Q - P` times the following is completed:
  - A number, `x`, is picked (any not previously picked from index `0` up to but not including index `Q`)
  - The question on that line of `_decrypting.txt` is provided for the person to answer.
  - The answer (as a UTF-8 string) is whitespace-trimmed, lowercased, and NFKC normalized.
  - This normalized answer is hashed with the SHA256 algorithm, and interpreted as a big endian big int as `y`.
  - `x` and `y` are stored as an XY point in the field with modulus `170,141,183,460,469,231,731,687,303,715,884,105,727` (the 127th Mersenne prime, referred to as M-127 from here on).
- The `P`  additional points Base64 are decoded as big endian big ints as `y` values, each having a corresponding `x` value equal to their line number.
  - Note: there will be a gap in the `x` coordinates corresponding to the line number of the explanatory comment.
- The [Lagrange polynomial](https://en.wikipedia.org/wiki/Lagrange_polynomial) is calculated from all `Q` points (both those generated from the selected questions, and from the additional points).
- The polynomial coefficients are used to calculate the missing `y` value at `x = Q` (using the same modulus).
- That big int is converted into (big endian) bytes, and encoded with Base64 (standard character set, no padding) — this is the password for the encrypted recipes.

<details>
<summary>A detailed worked example of decrypting</summary>

A `.protectedrecipes` file might include a `_decrypting.txt` file that looks like this:

```text _decrypting.txt
Look at the first recipe shown on page 81. In the recipe's instructions, what is the last word of the third step?
Look at the first recipe shown on page 22. How many people does this recipe cater for (as a number)?
Look at the first recipe shown on page 92. In the recipe's description, what is the second word of the third sentence?
Look at the second recipe shown on page 14. What is the recipe's title (including punctuation)?
Look at the first recipe shown on page 105. How many steps does this recipe have?
Look at the first recipe shown on page 123. In the recipe's instructions, what is the last word of the last step?
Look at the first recipe shown on page 66. In the recipe's description, what is the last word of the first sentence?
# Above are questions that allow the derivation of the password for the other files in this archive, below is additional machine information needed for the same. Please see https://github.com/jphastings/recipes#proof-of-ownership-extension for specifics.
S7yj6S74aoH+OHe13fSZyA
a5D0f5T4cNy6WsVHs9YXGw
F2rQyKGBPZiV58T8JYIq0Q
PJ+d02O5RqcAvSyHKkRlDw
```

As there are 7 question rows and 4 additional point rows this means `7 - 4 = 3` questions must be answered correctly to decrypt these recipes.

The decrypting process picks a random question to ask. The first is picked, so `x = 0 mod M-127` (as we're zero indexing):

> Look at the first recipe shown on page 81. In the recipe's instructions, what is the last word of the third step?

Assuming the answer is "flambé" then the SHA256 hash of that (downcased, whitespace trimmed, NKFC normalized) UTF-8 string is `0x46cf4447b9ee6996f564e9e5c2730f66a22e8ddd227e423c38fb77134476bfa7`, which means `y = 32028107995726419024310051710096036951208462986977060729683081194961185980327 mod M-127` (it is interpreted as a big endian decimal big int).

This process is repeated 2 more times (as three answers are needed), each `(x, y)` pair (modulus the M-127 prime) as a point.

Each of the 'additional point' rows is decoded with Base64 (standard, unpadded) and converted to a big endian big int, and added as a further point.

Line 8 (`x = 8 mod M-127`) is `S7yj6S74aoH+OHe13fSZyA` \
↓\
`y = 100671576000737126131661019523537213896 mod M-127`

All the points (from answers and the additional ones) are plugged into a Lagrange Interpolation Polynomial calculator. As the 7th line contains the comment the polynomial coefficients are used to find `y` when `x = 7` in order to derive the password.

Assuming the result is `y = 17047859747726120506192065393950404758 mod M-127` then this is converted to (big endian) bytes, and encoded as (unpadded, standard) Base64 characters, producing the password `DNNMYSmqMjqXG+HY0Lcwlg`.

If this password fails to decrypt the recipes then at least one of the answers is incorrect, a new question can be selected from the list and answered, and the process followed with each combination of 3 answers and all additional points until the correct password is found.

If no password has been correct and no questions remain to be selected, this means that all combinations of three answers have at least one incorrect answer within them, and the recipes cannot be decrypted — the person is assumed to not have a copy of the relevant physical book to hand.

</details>
