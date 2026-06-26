package ingredients

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/jphastings/recipes/internal/uuid"
)

type IngredientUse struct {
	Order      int        `json:"order"`
	UUID       uuid.UUID  `json:"uuid"`
	Quantity   Quantity   `json:"quantity"`
	Ingredient Ingredient `json:"ingredient"`
}

type Ingredient struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

type Quantity struct {
	Amount *Amount `json:"amount,omitempty"`
	Type   Unit    `json:"quantityType"`
}

type Unit string

type Amount big.Rat

func (a *Amount) MarshalJSON() ([]byte, error) {
	r := big.Rat(*a)
	f, _ := r.Float64()
	return json.Marshal(f)
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	r := new(big.Rat).SetFloat64(f)
	if r == nil {
		return fmt.Errorf("cannot represent %v as a rational amount", f)
	}
	*a = Amount(*r)
	return nil
}

var (
	UnitItem  Unit = "ITEM"
	UnitPinch Unit = "PINCH"
	UnitBunch Unit = "BUNCH"

	UnitTablespoon Unit = "TABLESPOON"
	UnitTeaspoon   Unit = "TEASPOON"
	UnitCup        Unit = "CUP"

	UnitMillilitre Unit = "MILLS"
	UnitCentilitre Unit = "CENTILITER"
	UnitDecilitre  Unit = "DECILITER"
	UnitLitre      Unit = "LITRES"

	UnitFluidOunce Unit = "FLUID_OUNCE"

	UnitGram     Unit = "GRAMS"
	UnitKilogram Unit = "KGS"

	UnitPound Unit = "POUND"
	UnitOunce Unit = "OUNCE"

	UnitBottle Unit = "BOTTLE"
	UnitCan    Unit = "CAN"
	UnitPacket Unit = "PACKET"

	SectionMarker Unit = "SECTION"
)

func (iu IngredientUse) String() string {
	return fmt.Sprintf("%s %s %s", (*big.Rat)(iu.Quantity.Amount), iu.Quantity.Type, iu.Ingredient.Name)
}

func NewSection(name string, order int) (IngredientUse, error) {
	uuid1, err1 := uuid.NewUUID("")
	uuid2, err2 := uuid.NewUUID("")
	if err1 != nil || err2 != nil {
		return IngredientUse{}, errors.Join(err1, err2)
	}

	return IngredientUse{
		Order:      order,
		Quantity:   Quantity{Type: SectionMarker},
		Ingredient: Ingredient{Name: name, UUID: uuid1},
		UUID:       uuid2,
	}, nil
}

// NewItem builds a plain, unit-less ingredient of quantity 1 named after the whole
// line. It is the fallback for ingredient text that ExtractIngredient cannot parse.
func NewItem(name string, order int) (IngredientUse, error) {
	iuUUID, err1 := uuid.NewUUID("")
	ingUUID, err2 := uuid.NewUUID(name)
	if err1 != nil || err2 != nil {
		return IngredientUse{}, errors.Join(err1, err2)
	}

	return IngredientUse{
		Order:      order,
		UUID:       iuUUID,
		Quantity:   Quantity{Amount: (*Amount)(big.NewRat(1, 1)), Type: UnitItem},
		Ingredient: Ingredient{Name: name, UUID: ingUUID},
	}, nil
}

func ExtractIngredient(ingredientLine string, order int) (IngredientUse, error) {
	ret, err := Parse("", []byte(ingredientLine))
	if err != nil {
		return IngredientUse{}, fmt.Errorf("couldn't parse '%s': %w", ingredientLine, err)
	}

	parts, ok := ret.([]any)
	if !ok {
		return IngredientUse{}, fmt.Errorf("couldn't determine parser response (was %T)", ret)
	}

	iuUUID, err := uuid.NewUUID("")
	if err != nil {
		return IngredientUse{}, err
	}

	amount, ok := parts[0].(*big.Rat)
	if !ok {
		amount = big.NewRat(1, 1)
	}

	unit, ok := parts[1].(Unit)
	if !ok {
		unit = UnitItem
	}

	name := parts[2].(string)
	ingUUID, err := uuid.NewUUID(name)
	if err != nil {
		return IngredientUse{}, err
	}

	return IngredientUse{
		Order: order,
		UUID:  iuUUID,
		Quantity: Quantity{
			Amount: (*Amount)(amount),
			Type:   unit,
		},
		Ingredient: Ingredient{
			Name: name,
			UUID: ingUUID,
		},
	}, nil
}
