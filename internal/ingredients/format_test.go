package ingredients_test

import (
	"math/big"
	"testing"

	"github.com/jphastings/recipes/internal/ingredients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func use(amount *big.Rat, unit ingredients.Unit, name string) ingredients.IngredientUse {
	return ingredients.IngredientUse{
		Quantity:   ingredients.Quantity{Amount: (*ingredients.Amount)(amount), Type: unit},
		Ingredient: ingredients.Ingredient{Name: name},
	}
}

// TestFormatIngredientUseReparses is the load-bearing invariant: every
// FormatIngredientUse output must parse back via ExtractIngredient to the same
// amount, unit, and name. The cases exercise every Unit in the mapping table and
// the integer / unicode-fraction / mixed / decimal amount renderings.
func TestFormatIngredientUseReparses(t *testing.T) {
	cases := []ingredients.IngredientUse{
		use(big.NewRat(1, 1), ingredients.UnitItem, "carrot"),
		use(big.NewRat(2, 1), ingredients.UnitItem, "carrots"),
		use(big.NewRat(7, 2), ingredients.UnitItem, "carrots"),  // 3½
		use(big.NewRat(1, 2), ingredients.UnitPinch, "salt"),    // ½
		use(big.NewRat(3, 2), ingredients.UnitPinch, "salt"),    // 1½
		use(big.NewRat(1, 3), ingredients.UnitBunch, "basil"),   // ⅓
		use(big.NewRat(8, 3), ingredients.UnitPacket, "yeast"),  // 2⅔
		use(big.NewRat(6, 5), ingredients.UnitKilogram, "beef"), // 1.2 (decimal fallback)
		use(big.NewRat(2, 1), ingredients.UnitTablespoon, "sugar"),
		use(big.NewRat(1, 1), ingredients.UnitTeaspoon, "vanilla"),
		use(big.NewRat(1, 1), ingredients.UnitCup, "flour"),
		use(big.NewRat(180, 1), ingredients.UnitMillilitre, "white wine"),
		use(big.NewRat(1, 1), ingredients.UnitCentilitre, "brandy"),
		use(big.NewRat(1, 1), ingredients.UnitDecilitre, "milk"),
		use(big.NewRat(1, 1), ingredients.UnitLitre, "stock"),
		use(big.NewRat(1, 1), ingredients.UnitFluidOunce, "white wine"),
		use(big.NewRat(11, 2), ingredients.UnitGram, "nutmeg"),
		use(big.NewRat(1, 1), ingredients.UnitPound, "potatoes"),
		use(big.NewRat(1, 1), ingredients.UnitOunce, "butter"),
		use(big.NewRat(1, 1), ingredients.UnitBottle, "wine"),
		use(big.NewRat(1, 1), ingredients.UnitCan, "tomatoes"),
	}

	for _, want := range cases {
		line := ingredients.FormatIngredientUse(want)
		t.Run(line, func(t *testing.T) {
			got, err := ingredients.ExtractIngredient(line, 0)
			require.NoError(t, err)

			assert.Equal(t, want.Ingredient.Name, got.Ingredient.Name)
			assert.Equal(t, want.Quantity.Type, got.Quantity.Type)
			wantAmt := (*big.Rat)(want.Quantity.Amount)
			gotAmt := (*big.Rat)(got.Quantity.Amount)
			assert.Zerof(t, wantAmt.Cmp(gotAmt), "amount: want %s got %s", wantAmt, gotAmt)
		})
	}
}
