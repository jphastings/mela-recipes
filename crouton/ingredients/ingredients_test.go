package ingredients_test

import (
	"math/big"
	"regexp"
	"testing"

	"github.com/jphastings/crouton-recipes/ingredients"
	"github.com/stretchr/testify/assert"
)

var isUUID = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")

func TestExtractIngredient(t *testing.T) {
	testCases := []struct {
		in     string
		name   string
		amount *big.Rat
		unit   ingredients.Unit
	}{
		{"a carrot", "carrot", big.NewRat(1, 1), ingredients.UnitItem},
		{"1 carrot", "carrot", big.NewRat(1, 1), ingredients.UnitItem},
		{"2 carrots", "carrots", big.NewRat(2, 1), ingredients.UnitItem},
		{"3.5 carrots", "carrots", big.NewRat(7, 2), ingredients.UnitItem},

		{"a pinch of salt", "salt", big.NewRat(1, 1), ingredients.UnitPinch},
		{"½ pinch of salt", "salt", big.NewRat(1, 2), ingredients.UnitPinch},
		{"1 pinch of salt", "salt", big.NewRat(1, 1), ingredients.UnitPinch},
		{"1½ pinches of salt", "salt", big.NewRat(3, 2), ingredients.UnitPinch},
		{"2 pinches of salt", "salt", big.NewRat(2, 1), ingredients.UnitPinch},
		{"2 pinch salt", "salt", big.NewRat(2, 1), ingredients.UnitPinch},

		{"half a bunch of basil", "basil", big.NewRat(1, 2), ingredients.UnitBunch},
		{"½ bunch of basil", "basil", big.NewRat(1, 2), ingredients.UnitBunch},
		{"⅓ bunch of basil", "basil", big.NewRat(1, 3), ingredients.UnitBunch},
		{"¼ bunch of basil", "basil", big.NewRat(1, 4), ingredients.UnitBunch},
		{"⅛ bunch of basil", "basil", big.NewRat(1, 8), ingredients.UnitBunch},
		{"⅔ bunch of basil", "basil", big.NewRat(2, 3), ingredients.UnitBunch},
		{"¾ bunch of basil", "basil", big.NewRat(3, 4), ingredients.UnitBunch},
		{"⅜ bunch of basil", "basil", big.NewRat(3, 8), ingredients.UnitBunch},
		{"⅝ bunch of basil", "basil", big.NewRat(5, 8), ingredients.UnitBunch},
		{"⅞ bunch of basil", "basil", big.NewRat(7, 8), ingredients.UnitBunch},
		{"½ a bunch of basil", "basil", big.NewRat(1, 2), ingredients.UnitBunch},
		{"1 bunch of basil", "basil", big.NewRat(1, 1), ingredients.UnitBunch},
		{"2 bunches of basil", "basil", big.NewRat(2, 1), ingredients.UnitBunch},

		{"2tbsp sugar", "sugar", big.NewRat(2, 1), ingredients.UnitTablespoon},
		{"2 tbsp sugar", "sugar", big.NewRat(2, 1), ingredients.UnitTablespoon},
		{"2 tbsps sugar", "sugar", big.NewRat(2, 1), ingredients.UnitTablespoon},

		{"3tsp sugar", "sugar", big.NewRat(3, 1), ingredients.UnitTeaspoon},
		{"3 tsp sugar", "sugar", big.NewRat(3, 1), ingredients.UnitTeaspoon},
		{"3 tsps sugar", "sugar", big.NewRat(3, 1), ingredients.UnitTeaspoon},

		{"1 cup sugar", "sugar", big.NewRat(1, 1), ingredients.UnitCup},
		{"3 cups sugar", "sugar", big.NewRat(3, 1), ingredients.UnitCup},

		{"180ml white wine", "white wine", big.NewRat(180, 1), ingredients.UnitMillilitre},
		{"180 mls white wine", "white wine", big.NewRat(180, 1), ingredients.UnitMillilitre},
		{"180 millilitres white wine", "white wine", big.NewRat(180, 1), ingredients.UnitMillilitre},

		{"20cl white wine", "white wine", big.NewRat(20, 1), ingredients.UnitCentilitre},
		{"30 cls white wine", "white wine", big.NewRat(30, 1), ingredients.UnitCentilitre},
		{"30 centilitres white wine", "white wine", big.NewRat(30, 1), ingredients.UnitCentilitre},

		{"2dl white wine", "white wine", big.NewRat(2, 1), ingredients.UnitDecilitre},
		{"3 dls white wine", "white wine", big.NewRat(3, 1), ingredients.UnitDecilitre},
		{"3 decilitres white wine", "white wine", big.NewRat(3, 1), ingredients.UnitDecilitre},

		{"1l white wine", "white wine", big.NewRat(1, 1), ingredients.UnitLitre},
		{"2.5 l white wine", "white wine", big.NewRat(25, 10), ingredients.UnitLitre},
		{"2.5 litres white wine", "white wine", big.NewRat(25, 10), ingredients.UnitLitre},

		{"1fl oz white wine", "white wine", big.NewRat(1, 1), ingredients.UnitFluidOunce},
		{"1fl. oz. white wine", "white wine", big.NewRat(1, 1), ingredients.UnitFluidOunce},
		{"1fl oz. white wine", "white wine", big.NewRat(1, 1), ingredients.UnitFluidOunce},
		{"1 floz white wine", "white wine", big.NewRat(1, 1), ingredients.UnitFluidOunce},
		{"1 fluid ounce white wine", "white wine", big.NewRat(1, 1), ingredients.UnitFluidOunce},
		{"2.5 fl oz white wine", "white wine", big.NewRat(25, 10), ingredients.UnitFluidOunce},
		{"2.5 fluid ounces white wine", "white wine", big.NewRat(25, 10), ingredients.UnitFluidOunce},

		{"5½g nutmeg", "nutmeg", big.NewRat(11, 2), ingredients.UnitGram},
		{"10g nutmeg", "nutmeg", big.NewRat(10, 1), ingredients.UnitGram},
		{"25 g nutmeg", "nutmeg", big.NewRat(25, 1), ingredients.UnitGram},
		{"25 grams nutmeg", "nutmeg", big.NewRat(25, 1), ingredients.UnitGram},

		{"½kg beef", "beef", big.NewRat(1, 2), ingredients.UnitKilogram},
		{"1.2kgs beef", "beef", big.NewRat(12, 10), ingredients.UnitKilogram},
		{"2 kilograms beef", "beef", big.NewRat(2, 1), ingredients.UnitKilogram},

		{"½oz flour", "flour", big.NewRat(1, 2), ingredients.UnitOunce},
		{"1 ounce flour", "flour", big.NewRat(1, 1), ingredients.UnitOunce},
		{"1.2 oz flour", "flour", big.NewRat(12, 10), ingredients.UnitOunce},
		{"2 ounces flour", "flour", big.NewRat(2, 1), ingredients.UnitOunce},

		{"½lb lamb", "lamb", big.NewRat(1, 2), ingredients.UnitPound},
		{"1 pound lamb", "lamb", big.NewRat(1, 1), ingredients.UnitPound},
		{"1.2lbs lamb", "lamb", big.NewRat(12, 10), ingredients.UnitPound},
		{"2 pounds lamb", "lamb", big.NewRat(2, 1), ingredients.UnitPound},

		{"1 bottle beer", "beer", big.NewRat(1, 1), ingredients.UnitBottle},
		{"2 bottles beer", "beer", big.NewRat(2, 1), ingredients.UnitBottle},

		{"half a can chopped tomatoes", "chopped tomatoes", big.NewRat(1, 2), ingredients.UnitCan},
		{"1 can chopped tomatoes", "chopped tomatoes", big.NewRat(1, 1), ingredients.UnitCan},
		{"2 cans chopped tomatoes", "chopped tomatoes", big.NewRat(2, 1), ingredients.UnitCan},

		{"1 packet yeast", "yeast", big.NewRat(1, 1), ingredients.UnitPacket},
		{"2⅔ packets yeast", "yeast", big.NewRat(8, 3), ingredients.UnitPacket},
	}

	for i, tc := range testCases {
		// This could be any number
		order := i + 1

		iu, err := ingredients.ExtractIngredient(tc.in, order)
		assert.NoError(t, err, tc.in)

		assert.Equal(t, tc.name, iu.Ingredient.Name, tc.in)
		assert.Equal(t, (*ingredients.Amount)(tc.amount), iu.Quantity.Amount, tc.in)
		assert.Equal(t, tc.unit, iu.Quantity.Type, tc.in)

		assert.Regexp(t, order, iu.Order, tc.in)
		assert.Regexp(t, isUUID, iu.UUID, tc.in)
		assert.Regexp(t, isUUID, iu.Ingredient.UUID, tc.in)
	}
}
