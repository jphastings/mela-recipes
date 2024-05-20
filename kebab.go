package mela

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var kebabCaser = regexp.MustCompile(`[^a-z0-9]+`)
var removeAccents = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func stringToFilename(str string) string {
	norm, _, _ := transform.String(removeAccents, str)
	return strings.Trim(
		kebabCaser.ReplaceAllString(
			strings.ReplaceAll(strings.ToLower(norm), "'", ""),
			"-"),
		"-")
}
