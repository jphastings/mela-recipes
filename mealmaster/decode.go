package mealmaster

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
)

// parseBlock maps one recipe block (header and terminator already stripped) into
// the interchange format. The block is three zones in order: labelled fields
// (Title/Categories/Yield), a fixed-column ingredient list, and free-text
// directions.
func parseBlock(lines []string) (formats.InterchangeRecipe, error) {
	ir := formats.NewInterchangeRecipe()

	i := 0
	for ; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			continue
		}
		label, value, ok := labelledField(t)
		if !ok {
			break // first non-field line begins the ingredient/direction body
		}
		switch label {
		case "title":
			ir.Title = value
		case "categories":
			ir.Tags = splitCategories(value)
		case "yield":
			ir.Yield = value
		}
	}

	groups, directions := parseBody(lines[i:])
	ir.Ingredients = groups
	if len(directions) > 0 {
		ir.Instructions = []formats.TitledList{{List: directions}}
	}

	if ir.Title == "" {
		return formats.InterchangeRecipe{}, errors.New("MealMaster recipe has no title")
	}
	return ir, nil
}

// parseBody splits the lines after the labelled fields into structured ingredient
// groups and free-text direction paragraphs. The ingredient block comes first;
// the first line that isn't an ingredient (or section header / continuation)
// starts the directions.
func parseBody(lines []string) ([]formats.IngredientGroup, []string) {
	var groups []formats.IngredientGroup
	cur := formats.IngredientGroup{}
	order := 0

	flush := func() {
		if cur.Title != "" || len(cur.Items) > 0 {
			groups = append(groups, cur)
			cur = formats.IngredientGroup{}
		}
	}

	i := 0
	for ; i < len(lines); i++ {
		line := lines[i]
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		if title, ok := sectionHeader(t); ok {
			flush()
			cur.Title = title
			continue
		}

		if left, right, ok := twoColumn(line); ok {
			appendColumn(&cur, left, &order)
			appendColumn(&cur, right, &order)
			continue
		}

		amount, unit, text, ok := parseIngredientColumns(line)
		if !ok {
			break // directions start here
		}
		if amount == "" && unit == "" && strings.HasPrefix(text, "-") && len(cur.Items) > 0 {
			appendContinuation(&cur.Items[len(cur.Items)-1], strings.TrimSpace(text[1:]))
			continue
		}
		cur.Items = append(cur.Items, buildIngredient(amount, unit, text, order))
		order++
	}
	flush()

	return groups, collectDirections(lines[i:])
}

// appendColumn parses one already-split single-column segment and appends it to
// the group.
func appendColumn(g *formats.IngredientGroup, segment string, order *int) {
	amount, unit, text, ok := parseIngredientColumns(segment)
	if !ok {
		return
	}
	g.Items = append(g.Items, buildIngredient(amount, unit, text, *order))
	*order++
}

// labelledField recognises the `Title:`, `Categories:`, and `Yield:` header
// fields, returning the lower-cased label and its trimmed value.
func labelledField(t string) (label, value string, ok bool) {
	i := strings.IndexByte(t, ':')
	if i < 0 {
		return "", "", false
	}
	label = strings.ToLower(strings.TrimSpace(t[:i]))
	switch label {
	case "title", "categories", "yield":
		return label, strings.TrimSpace(t[i+1:]), true
	}
	return "", "", false
}

// splitCategories turns a comma-separated category list into tags, dropping the
// MealMaster "None" placeholder.
func splitCategories(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" && !strings.EqualFold(t, "none") {
			out = append(out, t)
		}
	}
	return out
}

var sectionRE = regexp.MustCompile(`^(?:MMMMM)?-{3,}\s*(.+?)\s*-{3,}$`)

// sectionHeader extracts the title from a hyphen-wrapped ingredient section
// header (e.g. `-----Sauce-----` or `MMMMM-----Sauce-----`).
func sectionHeader(t string) (string, bool) {
	m := sectionRE.FindStringSubmatch(t)
	if m == nil {
		return "", false
	}
	if title := strings.TrimSpace(m[1]); title != "" {
		return title, true
	}
	return "", false
}

// parseIngredientColumns reads a single-column MealMaster ingredient line by its
// fixed layout: amount in columns 1–7, a 2-letter unit code in columns 9–10, and
// the ingredient text from column 12. It rejects lines that don't fit that shape
// (the discriminator between the ingredient block and free-text directions).
func parseIngredientColumns(line string) (amount, unit, text string, ok bool) {
	r := []rune(strings.TrimRight(line, " "))
	if len(r) < 12 || r[7] != ' ' || r[10] != ' ' {
		return "", "", "", false
	}

	amount = strings.TrimSpace(string(r[0:7]))
	unit = strings.TrimSpace(string(r[8:10]))
	text = strings.TrimSpace(string(r[11:]))

	switch {
	case text == "":
		return "", "", "", false
	case !quantityOrBlank(amount) || !lettersOrBlank(unit):
		return "", "", "", false
	case amount == "" && unit == "" && !strings.HasPrefix(text, "-"):
		// A bare line with no amount or unit is almost certainly a direction, not
		// an ingredient (a continuation line is the one exception).
		return "", "", "", false
	}
	return amount, unit, text, true
}

// twoCol is the width of one ingredient column in MealMaster's side-by-side
// layout; the second column starts at column 41.
const twoCol = 40

// twoColumn detects the two-column ingredient layout and splits a physical line
// into its two single-column halves, but only when both halves parse as
// ingredients (so a long single-column ingredient name isn't split by mistake).
func twoColumn(line string) (left, right string, ok bool) {
	r := []rune(line)
	if len(r) <= twoCol {
		return "", "", false
	}
	left = string(r[:twoCol])
	right = string(r[twoCol:])
	if _, _, _, lok := parseIngredientColumns(left); !lok {
		return "", "", false
	}
	if _, _, _, rok := parseIngredientColumns(right); !rok {
		return "", "", false
	}
	return left, right, true
}

// buildIngredient reconstructs a grammar-friendly "amount unit text" line —
// converting MealMaster's ASCII fractions and 2-letter unit codes — and parses it
// into a structured ingredient.
func buildIngredient(amount, unit, text string, order int) ingredients.IngredientUse {
	var parts []string
	if a := ingredients.NormalizeAmount(amount); a != "" {
		parts = append(parts, a)
	}
	if u := expandUnit(unit); u != "" {
		parts = append(parts, u)
	}
	parts = append(parts, text)
	return ingredients.ParseOrItem(strings.Join(parts, " "), order)
}

// appendContinuation folds a continuation line ("-" in the text column) onto the
// previous ingredient's name.
func appendContinuation(iu *ingredients.IngredientUse, extra string) {
	if extra == "" {
		return
	}
	iu.Ingredient.Name = strings.TrimSpace(iu.Ingredient.Name + " " + extra)
}

// collectDirections joins the wrapped lines of each direction paragraph (blank
// lines separate paragraphs) into one instruction per paragraph.
func collectDirections(lines []string) []string {
	var paragraphs []string
	var current []string
	flush := func() {
		if len(current) > 0 {
			paragraphs = append(paragraphs, strings.Join(current, " "))
			current = nil
		}
	}

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			flush()
			continue
		}
		if isTerminator(line) {
			continue
		}
		current = append(current, t)
	}
	flush()

	return paragraphs
}

func quantityOrBlank(s string) bool {
	for _, r := range s {
		if !strings.ContainsRune("0123456789./- ", r) {
			return false
		}
	}
	return true
}

func lettersOrBlank(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

// unitCodes maps MealMaster's 2-letter unit abbreviations to a word the
// ingredient grammar understands; codes mapped to "" (each, per, small, large)
// drop to a bare count. Unmapped codes are kept verbatim by expandUnit.
var unitCodes = map[string]string{
	"c": "cup", "C": "cup",
	"t": "tsp", "ts": "tsp",
	"T": "tbsp", "tb": "tbsp", "tbs": "tbsp",
	"oz": "oz", "fl": "fl oz",
	"g": "g", "kg": "kg", "mg": "mg",
	"lb": "lb", "lbs": "lb",
	"ml": "ml", "l": "l", "cl": "cl", "dl": "dl",
	"cn": "can", "pk": "packet", "pkg": "packet", "pg": "packet",
	"pn": "pinch", "ds": "dash", "dr": "drop",
	"qt": "quart", "pt": "pint", "ga": "gallon", "gl": "gallon",
	"sl": "slice", "dz": "dozen", "bn": "bunch",
	"ea": "", "x": "", "sm": "", "lg": "",
}

func expandUnit(code string) string {
	c := strings.TrimSpace(code)
	if c == "" {
		return ""
	}
	if word, ok := unitCodes[c]; ok {
		return word
	}
	if word, ok := unitCodes[strings.ToLower(c)]; ok {
		return word
	}
	return c
}
