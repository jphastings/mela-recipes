package ingredients

import (
	"math/big"
	"strings"
)

// unitWords maps each Unit to the canonical short form the grammar's Unit rule
// re-accepts. UnitItem, SectionMarker and any unknown unit map to "" so no unit
// word is emitted (ExtractIngredient defaults a missing unit to UnitItem).
var unitWords = map[Unit]string{
	UnitPinch:      "pinch",
	UnitBunch:      "bunch",
	UnitTablespoon: "tbsp",
	UnitTeaspoon:   "tsp",
	UnitCup:        "cup",
	UnitMillilitre: "ml",
	UnitCentilitre: "cl",
	UnitDecilitre:  "dl",
	UnitLitre:      "l",
	UnitFluidOunce: "fl oz",
	UnitGram:       "g",
	UnitKilogram:   "kg",
	UnitPound:      "lb",
	UnitOunce:      "oz",
	UnitBottle:     "bottle",
	UnitCan:        "can",
	UnitPacket:     "packet",
}

// fractionGlyphs maps a reduced fractional part (as big.Rat.RatString gives it) to
// the exact unicode glyph the grammar's Fraction rule recognises.
var fractionGlyphs = map[string]string{
	"1/2": "½",
	"1/3": "⅓",
	"1/4": "¼",
	"1/8": "⅛",
	"2/3": "⅔",
	"3/4": "¾",
	"3/8": "⅜",
	"5/8": "⅝",
	"7/8": "⅞",
}

// FormatIngredientUse renders iu into a string that ExtractIngredient parses back
// into the same amount, unit, and name — it is the inverse of grammar.peg. Section
// markers are not rendered here: callers detect Quantity.Type == SectionMarker and
// use the ingredient name as a heading instead.
func FormatIngredientUse(iu IngredientUse) string {
	var parts []string
	if qty := RenderQuantity(iu); qty != "" {
		parts = append(parts, qty)
	}
	parts = append(parts, iu.Ingredient.Name)
	line := strings.Join(parts, " ")
	if iu.Note != "" {
		line += " (" + iu.Note + ")"
	}
	return line
}

// RenderQuantity renders the "amount unit" portion of an ingredient (eg. "2 g",
// "1½", "1 tbsp", "2 cups"). When no amount was given and there is no unit — the
// shape an unquantified line parses to — it returns "" so callers emit just the
// name rather than a spurious "1".
func RenderQuantity(iu IngredientUse) string {
	amount := (*big.Rat)(iu.Quantity.Amount)
	word := unitWords[iu.Quantity.Type]
	switch {
	case amount == nil && word == "":
		return ""
	case word == "":
		return renderAmount(amount)
	case amount == nil:
		return word
	default:
		return renderAmount(amount) + " " + word
	}
}

// UnitWord returns the canonical short word for a unit ("g", "tbsp", …), or ""
// for units that aren't rendered (UnitItem, SectionMarker, unknown).
func UnitWord(u Unit) string {
	return unitWords[u]
}

// renderAmount renders r using the most exact form the parser accepts: a whole
// number, a whole number plus a unicode fraction, a bare unicode fraction, or a
// terminating decimal. A nil amount is treated as 1, matching ExtractIngredient
// which defaults a missing amount to 1. An amount that is neither a supported
// fraction nor a terminating decimal (e.g. 1/7) falls back to its closest 6-dp
// decimal and is lossy for that single amount — no worse than crouton's own
// float64 on-disk representation.
func renderAmount(r *big.Rat) string {
	if r == nil {
		return "1"
	}
	if r.IsInt() {
		return r.Num().String()
	}

	whole := new(big.Int).Quo(r.Num(), r.Denom())
	rem := new(big.Rat).Sub(r, new(big.Rat).SetInt(whole))
	if glyph, ok := fractionGlyphs[rem.RatString()]; ok {
		if whole.Sign() == 0 {
			return glyph
		}
		return whole.String() + glyph
	}

	dec := strings.TrimRight(strings.TrimRight(r.FloatString(6), "0"), ".")
	if parsed, ok := new(big.Rat).SetString(dec); ok && parsed.Cmp(r) == 0 {
		return dec
	}
	return r.FloatString(6)
}
