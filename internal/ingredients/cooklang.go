package ingredients

import (
	"math/big"

	"github.com/jphastings/recipes/internal/uuid"
)

// unitByWord inverts unitWords so a Cooklang unit token (e.g. "g", "fl oz") can
// be mapped back to its Unit. The unitWords values are all distinct, so this is
// an exact inverse for every token CooklangQuantity emits.
var unitByWord = func() map[string]Unit {
	m := make(map[string]Unit, len(unitWords))
	for u, w := range unitWords {
		m[w] = u
	}
	return m
}()

// CooklangQuantity renders iu's amount and unit as the two halves of a Cooklang
// quantity body ("{amount%unit}"). unit is "" for unitless ingredients, and both
// are "" when the ingredient is a single unitless item (amount 1), so the caller
// can emit a bare "@name" with no braces. It is the inverse of FromCooklang.
func CooklangQuantity(iu IngredientUse) (amount, unit string) {
	unit = unitWords[iu.Quantity.Type]
	amount = renderAmount((*big.Rat)(iu.Quantity.Amount))
	if unit == "" && amount == "1" {
		return "", ""
	}
	return amount, unit
}

// FromCooklang builds an IngredientUse from the parts of a Cooklang ingredient
// reference — the quantity amount ("500", "½", "3½", or "" for none), the unit
// token ("g", "tbsp", or "" for none), and the name. It is the parse-side inverse
// of CooklangQuantity: an unparseable amount falls back to 1, and an unrecognised
// unit word is folded into the name so no authored text is silently dropped.
func FromCooklang(amount, unit, name string, order int) (IngredientUse, error) {
	amt := big.NewRat(1, 1)
	if amount != "" {
		if probe, err := ExtractIngredient(amount+" x", 0); err == nil && probe.Quantity.Amount != nil {
			amt = (*big.Rat)(probe.Quantity.Amount)
		}
	}

	unitType := UnitItem
	if unit != "" {
		if u, ok := unitByWord[unit]; ok {
			unitType = u
		} else {
			name = unit + " " + name
		}
	}

	iuUUID, err1 := uuid.NewUUID("")
	ingUUID, err2 := uuid.NewUUID(name)
	if err1 != nil {
		return IngredientUse{}, err1
	}
	if err2 != nil {
		return IngredientUse{}, err2
	}

	return IngredientUse{
		Order:      order,
		UUID:       iuUUID,
		Quantity:   Quantity{Amount: (*Amount)(amt), Type: unitType},
		Ingredient: Ingredient{Name: name, UUID: ingUUID},
	}, nil
}
