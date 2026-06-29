package ingredients

import (
	"math/big"
	"strings"
)

// asciiFractionGlyphs maps the common ASCII fractions to the unicode glyphs the
// grammar's Fraction rule accepts.
var asciiFractionGlyphs = map[string]string{
	"1/2": "½", "1/3": "⅓", "2/3": "⅔", "1/4": "¼", "3/4": "¾",
	"1/8": "⅛", "3/8": "⅜", "5/8": "⅝", "7/8": "⅞",
}

// NormalizeAmount converts a free-text quantity that uses ASCII fractions (eg.
// "1/2", "1 1/2") into a form the ingredient grammar accepts: common fractions
// become unicode glyphs, mixed numbers fold into a single glyph or a decimal, and
// plain integers/decimals pass through unchanged. It is for importers that
// reconstruct an ingredient line from a format's separate quantity field.
func NormalizeAmount(raw string) string {
	fields := strings.Fields(raw)
	switch len(fields) {
	case 0:
		return ""
	case 1:
		return convertFraction(fields[0])
	case 2:
		if glyph, ok := asciiFractionGlyphs[fields[1]]; ok && isInteger(fields[0]) {
			return fields[0] + glyph
		}
		if whole, ok := new(big.Rat).SetString(fields[0]); ok {
			if frac, ok := new(big.Rat).SetString(fields[1]); ok {
				return ratDecimal(new(big.Rat).Add(whole, frac))
			}
		}
	}
	return raw
}

func convertFraction(t string) string {
	if glyph, ok := asciiFractionGlyphs[t]; ok {
		return glyph
	}
	if strings.ContainsRune(t, '/') {
		if r, ok := new(big.Rat).SetString(t); ok {
			return ratDecimal(r)
		}
	}
	return t
}

func ratDecimal(r *big.Rat) string {
	s := r.FloatString(4)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}

func isInteger(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
