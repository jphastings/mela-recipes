- [ ] Add `SourceFilename()` to `formats.Recipe` so I can always report on where the recipe has come from?
- [ ] Mela: preserve `Notes` when importing from another format — `mela.ImportRecipe` currently drops the interchange `Notes`, so converting e.g. Paprika → Mela loses the notes text (the book reference still survives via the ID).
- [x] Implement single-recipe output in `cmd/recipes` — `makeRecipes` imports each recipe and writes it to a file, honouring `--out-here`/`--out-there`/`--out-dir`, `--overwrite`, and an explicit `--to <file>.<ext>`.
- [x] Build a tool for downloading remote images, so URL-only images (e.g. Paprika's `image_url`) can be fetched into a `B64Image` during conversion. (`utils.FetchImage`, used by schema.org and Paprika, gated behind `--network` / `ParseOptions.AllowNetwork`)

## Lossy-transfer semantic tagging

- [ ] Design lossy-transfer semantic tagging: a way for a conversion to record and report which fields couldn't be represented in the target format, instead of dropping them silently.

Once that exists, surface each of these known losses through it:

- [ ] Paprika import keeps only the first image in `photo_data` — tag any further images as dropped.
- [ ] Paprika `image_url` is now fetched when `--network` is set; when it isn't, tag it as an unresolved remote image (lossy-transfer).
- [ ] Paprika `categories` are treated as plain name strings, whereas the Paprika cloud API uses category UUIDs — tag the ambiguity.
- [ ] Paprika-only fields with no interchange equivalent (`nutritional_info`, `source`, `source_url`, `rating`, `difficulty`) are dropped on cross-format conversion — tag them.
- [ ] Mela drops `link` and `nutrition` when exporting to the interchange format — tag them.
