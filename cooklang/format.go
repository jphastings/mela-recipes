package cooklang

import (
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
)

const (
	recipeExt = ".cook"
)

var FormatInfo = &formats.Format{
	Name: "Cooklang",
	URL:  "https://cooklang.org",
	Features: formats.Features{
		ParseRecipe: true,
		WriteRecipe: true,
	},
	Lossiness: formats.Lossiness{
		// Interchange -> Cooklang reparses each ingredient line into a structured
		// quantity so it can be written as inline @markup; wording that doesn't
		// survive that parse-then-format is normalised, and an ingredient inlined
		// into a step takes the position its name occupies in the prose.
		OnImport: []formats.LossyField{
			{
				Field:   "Ingredients",
				Reason:  "ingredient wording is normalised (eg. '2 tablespoons' becomes '2 tbsp'), and ingredients inlined into steps follow the order their names appear in the instructions",
				Present: hasLossyIngredientText,
			},
		},
		// Cooklang -> interchange has nowhere to keep cookware or timers, so their
		// #/~ markup is folded into the step text; comments are dropped.
		OnExport: []formats.LossyField{
			{
				Field:  "Instructions",
				Reason: "cookware (#) and timers (~) become plain words in the step text, and comments are dropped — they cannot be re-emitted as markup",
			},
		},
	},
	Extension: recipeExt,
	Import:    importRecipe,
	Parse:     Parse,
	Bundle:    bundle,
}

// hasLossyIngredientText reports whether any ingredient line would not survive
// the structured parse-then-format round-trip unchanged (eg. "a handful of
// basil" becomes "1 handful of basil", "2 tablespoons" becomes "2 tbsp").
func hasLossyIngredientText(ir formats.InterchangeRecipe) bool {
	for _, tl := range ir.Ingredients {
		for _, line := range tl.List {
			iu, err := ingredients.ExtractIngredient(line, 0)
			if err != nil || ingredients.FormatIngredientUse(iu) != line {
				return true
			}
		}
	}
	return false
}

var bundleExts = []string{recipeExt, ".jpg", ".jpeg", ".png"}
var sectionSuffix = regexp.MustCompile(`\.\d+$`)

func bundle(files []string) (bundles []formats.Bundle, unused []string) {
	idx := make(map[string][]string)

	for _, f := range files {
		ext := path.Ext(f)
		if !slices.Contains(bundleExts, ext) {
			unused = append(unused, f)
			continue
		}
		k := strings.TrimSuffix(f, ext)
		if ext != recipeExt {
			k = sectionSuffix.ReplaceAllString(k, "")
		}
		idx[k] = append(idx[k], f)
	}

	for _, b := range idx {
		bundles = append(bundles, b)
	}
	return
}
