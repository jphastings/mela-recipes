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

If the URN _also_ includes a valid `recipe` f-component, then the recipe will be interpreted as coming from the Nth recipe on the first page referenced in `pages`. `0` represents "not explicitly specified, presumed the first recipe", `1` explicitly declares this recipe as the first one on the page, `2` explicitly as the second and so on. (Neatly resolving the awkward difference between humans and machines on zero-indexing).

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

Files protected like this are identical to `.melarecipes` zip files, except that they use the extension `.protectedrecipes`, all contained `.melarecipe` files are password protected with standard AES256 zip encryption, and contain a `decrypting.txt` (in the clear) with the questions that need to be answered to derive the decryption password.

The `decrypting.txt` file has seven questions on seven separate lines (written in the locale of the book in question, `\n` delimited), and expecting a one-word answer. (The file also ends in a `#` prefixed comment pointing to this readme).

To prove ownership at least 4 of the 7 questions must be answered, and processed as follows to create the password:

- Each answer (as a UTF-8 string) is whitespace-trimmed, lowercased, and NFKC normalized
- This is hashed with the CRC32 algorithm
- Bit `0` from the CRC32 of answer `i` (of 7) is used as bit `i` of a [Hamming(7,4) code](https://en.wikipedia.org/wiki/Hamming(7,4)), which become the high nibble of the first byte of the unencoded password
- Bit `1` from the CRC32 follows the same process to become the low nibble of the first byte of the unencoded password
- The resulting 16 bytes of the unencoded password become 22 bytes of (unpadded) Base64 characters, which is the password used for the recipes

#### Decrypting example

A `.protectedrecipes` file might include a `_decrypting.txt` file that looks like this:

<summary>
<details>Example `_decrypting.txt`</details>

```text
Look at the first recipe shown on page 81. In the recipe's instructions, what is the last word of the third step?
Look at the first recipe shown on page 22. How many people does this recipe cater for (as a number)?
Look at the first recipe shown on page 92. In the recipe's description, what is the second word of the third sentence?
Look at the second recipe shown on page 14. What is the recipe's title (including punctuation)?
Look at the (…etc for Q5)
Look at the (…etc for Q6)
Look at the (…etc for Q7)
# These are questions that allow the derivation of the password for the other files in this archive. Please see https://github.com/jphastings/mela-recipes#proof-of-ownership-extension for information.
```

</summary>

The decrypting process picks a random question to ask (the one at index 0):

> Look at the first recipe shown on page 81. In the recipe's instructions, what is the last word of the third step?

Assuming the answer is "transparent" then the CRC32 of that (downcased, whitespace trimmed, NKFC normalized) UTF-8 string is `0x0875e90c`, or `0b00001000011101011110100100001100`.

Similarly the CRC32 hashes of the questions with index 3, 5, and 6 are obtained:

| Index    | CRC32 of normalized answer         |
|----------|------------------------------------|
| 0 ✅ (p1) | `00001000011101011110100100001100` |
| 1 ❌ (p2) | `????????????????????????????????` |
| 2 ❌ (d1) | `????????????????????????????????` |
| 3 ✅ (p3) | `01010011101101011010001111001110` |
| 4 ❌ (d2) | `????????????????????????????????` |
| 5 ✅ (d3) | `11100000001011101111011110011001` |
| 6 ✅ (d4) | `01001010110100000110000000010100` |

The first column is treated as a Hamming(7,4) code, with the parity bits used to calculate any missing data bits (if any of questions 2, 4, 5 and 6 weren't selected to be answered).

`0??0?10` → `1110` (or `0xE`) because:

- `p3 = d2 + d3 + d4 % 2`
  `0  = ?  + 1  + 0  % 2`
  `d2 = 1`
- `p1 = d1 + d2 + d4 % 2`
  `0  = ?  + 1  + 0  % 2`
  `d1 = 1`

Completing this process for the remaining columns, the unencoded password (expressed in hex here) is `0xE7ECD09C9DA1EAE4A3FE8E2A2C0BA1CE`. Expressed in Base64 (without padding) is `5+zQnJ2h6uSj/o4qLAuhzg` — the password for this `.protectedrecipes` zip file.

If the password used fails to decrypt any recipe, it can be assumed that one (or more) of the given answers was incorrectly provided. Up to three supplementary questions can be asked to attempt to correct for up to one incorrectly provided answer.
