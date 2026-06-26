# protected

Reads and writes **protected archives**: zip files whose entries are AES256
encrypted with a password that can only be reconstructed by answering questions
about their contents.

The intended use is bundling recipes derived from a physical book such that only
someone able to consult that book — and so, presumably, an owner of it — can
decrypt them. The questions are generated from each recipe's interchange
representation, so the package works with **any recipe type**.

## How it works

A protected archive is an ordinary zip containing:

- `_decrypting.txt` (unencrypted, first entry) — the **manifest**: the human
  answerable questions, then a comment line, then "additional points".
- One AES256-encrypted entry per recipe (in its native format, eg. `.melarecipe`).

The decryption password is shared across the questions using
[Lagrange interpolation](https://en.wikipedia.org/wiki/Lagrange_polynomial) over
the field of the 127th Mersenne prime. Each correct answer recovers one point on a
polynomial; the additional points in the manifest supply the rest, so only a
subset of the questions need be answered correctly. The password is the
polynomial evaluated at a reserved coordinate. See the
[format specification](../mela/README.md#proof-of-ownership-extension) for the
on-disk details and a worked example.

## Usage

Writing — add any `formats.Recipe`s and the questions are generated on `Close`:

```go
err := protected.Create("book.protectedrecipes", recipes)
// or, incrementally:
w := protected.NewWriter(f, protected.WithRequiredCorrect(3))
for _, r := range recipes {
    w.Add(r)
}
err := w.Close()
```

At least `WithQuestionCount` (default 8) suitable recipes must be added, of which
`WithRequiredCorrect` (default 4) questions must later be answered correctly.

Reading — supply callbacks that present questions and collect answers:

```go
entries, closer, err := protected.Open("book.protectedrecipes", onTest, onExplain)
defer closer.Close()
for _, e := range entries {
    rc, _ := e.Open()        // decrypts on demand
    // parse rc in the appropriate format, eg. mela.ParseRecipeStream(rc)
    rc.Close()
}
```

`onTest` is called once per question and returns the typed answer (normalised
before use); returning an empty answer skips that question. `onExplain` is called
once beforehand with how many answers are needed and how many may be skipped.
