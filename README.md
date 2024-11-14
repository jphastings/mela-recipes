# Recipes

Tools and packages for parsing, building, and converting (cookbook) recipe formats.

- Read & write [Mela](https://mela.recipes)'s `.melarecipe` and `.melarecipes` files
- Read & write [Crouton](https://crouton.app/)'s `.crumb` files
- Read & write [Paprika](https://www.paprikaapp.com/)'s `.paprikarecipe` and `.paprikarecipes` files
- Read & write [Cooklang](https://cooklang.org/)'s `.cook` files
- Convert _between_ these formats
- Extract recipe files from cookbook ePubs
- Libraries for all the above in your own projects

## Usage

The `recipes` CLI tool has multiple commands for interacting with recipe files.

```bash
# Convert a Mela recipe to a Crouton recipe with the same name
$ recipes convert --to crouton mums-lasagne.melarecipe
🗂️ Converting 1 recipe file
📖 Found Mums Lasagne at mums-lasagne.melarecipe…
  💾 …saved to mums-lasagne.crumb
  ⚠️ Recipe description not supported in Crouton recipes
```

```bash
# Convert and add all recipes in a folder to a named Mela recipes collection 
$ recipes convert --to baked-goods.melarecipes baking/*
🗂️ Converting 3 recipe files
🧑‍🍳 Found Bakewell Tart at baking/bakewell-tart.melarecipe…
  📥 …added into baked-goods.melarecipes
🧑‍🍳 Found Lemon Merengue at baking/lemon-merengue.crumb…
  📥 …added into baked-goods.melarecipes
🧑‍🍳 Found Chocolate hazelnut brownies at baking/chocolate-hazelnut-brownies.melarecipe…
  📥 …added into baked-goods.melarecipes
📔 3 Recipes found
📦 Finished adding to baked-goods.melarecipes
```

```bash
# Search for recipes in an ePub, creating a Crouton recipe for each as well as adding all to a Mela recipes collection
$ recipes convert --to .melarecipes,.crumb 9780451496614.epub
📖 Extracting recipes from The Palomar Cookbook
🧑‍🍳 Found Cured Lemons…
  📥 …added into the-palomar-cookbook.melarecipes
  💾 …saved to the-palomar-cookbook/cured-lemons.crumb
  ⚠️ Recipe description not supported in Crouton recipes
🧑‍🍳 Found Cured Lemon Paste…
  📥 …added to the-palomar-cookbook.melarecipes
  💾 …saved to the-palomar-cookbook/cured-lemon-paste.crumb
  ⚠️ Recipe description not supported in Crouton recipes
📔 91 Recipes found in The Palomar Cookbook
📦 Finished adding to the-palomar-cookbook.melarecipes
```

## Extensions

This code also includes some extensions to the various recipe file formats:

- **ISBNs**: storing information about the physical books recipes are digitised from (ISBN, page number, and recipe-on-page number).
- **Image optimize**: pixel- and byte-size reduction of images in recipes to a maximum width and height (512×512px by default) and with excellent [JPEGli](https://opensource.googleblog.com/2024/04/introducing-jpegli-new-jpeg-coding-library.html) compression.

## Feature Support

Different recipe formats have support for different functionality. This table shows where information may be lost on conversion, or additional supported features. Features are fully supported (✅), supported via an extension (⚙️) or unsupported (❌).

| Feature            | Mela | Crouton | Paprika | Cooklang |
|--------------------|:----:|:-------:|:-------:|:--------:|
| Image optimize     |  ✅   |    ✅    |    ❔    |    ✅     |
| Recipe description |  ✅   |    ❌    |    ✅    |    ⚙️     |
| ISBNs              |  ⚙️   |    ❌    |    ❔    |    ⚙️     |
