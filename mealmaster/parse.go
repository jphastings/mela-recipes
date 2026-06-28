package mealmaster

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/jphastings/recipes/internal/formats"
	"golang.org/x/text/encoding/charmap"
)

// Parse reads a single `.mmf` file — which commonly concatenates many recipes —
// and streams one ParseEvent per recipe found, with a running I and the total N.
// A `.mmf` is a flat archive rather than a named app collection, so no
// CollectionDetails are returned.
func Parse(b formats.Bundle, _ formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	filename := b[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	blocks := splitRecipes(decodeText(data))

	pe := make(chan formats.ParseEvent, len(blocks)+1)
	go func() {
		defer close(pe)
		pe <- formats.ParseEvent{N: len(blocks)}

		if len(blocks) == 0 {
			pe <- formats.ParseEvent{Err: fmt.Errorf("%s: no MealMaster recipes found", filename)}
			return
		}

		for _, block := range blocks {
			ir, err := parseBlock(block)
			if err != nil {
				pe <- formats.ParseEvent{Err: fmt.Errorf("%s: %w", filename, err), I: 1}
				continue
			}
			pe <- formats.ParseEvent{Recipe: &Recipe{ir: ir}, I: 1}
		}
	}()

	return pe, nil, nil
}

// decodeText returns the file as UTF-8 text. MealMaster archives are usually
// CP437 (DOS) or already UTF-8; bytes that are valid UTF-8 are kept as-is,
// otherwise they're decoded from CP437 (so fractions and degree signs survive).
func decodeText(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	if decoded, err := charmap.CodePage437.NewDecoder().Bytes(data); err == nil {
		return string(decoded)
	}
	return string(data)
}

// splitRecipes carves the file into per-recipe blocks. A recipe begins at a
// header line (`-----`/`MMMMM` marker mentioning Meal-Master) and runs until the
// next terminator (a line of only hyphens, or `MMMMM`) or the next header. The
// returned blocks exclude the header and terminator lines.
func splitRecipes(text string) [][]string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		lines = append(lines, strings.TrimRight(line, "\r"))
	}

	var blocks [][]string
	for i := 0; i < len(lines); {
		if !isHeader(lines[i]) {
			i++
			continue
		}
		i++ // skip the header line

		var body []string
		for i < len(lines) && !isTerminator(lines[i]) && !isHeader(lines[i]) {
			body = append(body, lines[i])
			i++
		}
		if i < len(lines) && isTerminator(lines[i]) {
			i++ // consume the terminator
		}
		blocks = append(blocks, body)
	}
	return blocks
}

// isHeader reports whether a line is a recipe's "via Meal-Master" header marker.
func isHeader(line string) bool {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "-----") && !strings.HasPrefix(t, "MMMMM") {
		return false
	}
	low := strings.ToLower(t)
	return strings.Contains(low, "meal-master") || strings.Contains(low, "mealmaster")
}

// isTerminator reports whether a line ends a recipe: a run of only hyphens, or
// `MMMMM` optionally followed by only hyphens. Hyphen-wrapped lines that contain
// text are ingredient section headers, not terminators.
func isTerminator(line string) bool {
	t := strings.TrimSpace(line)
	switch {
	case t == "":
		return false
	case len(t) >= 5 && strings.Trim(t, "-") == "":
		return true
	case strings.HasPrefix(t, "MMMMM") && strings.Trim(strings.TrimPrefix(t, "MMMMM"), "-") == "":
		return true
	default:
		return false
	}
}
