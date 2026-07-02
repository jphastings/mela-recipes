package cooklang

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
	"gopkg.in/yaml.v3"
)

// marshalCook renders an interchange recipe as a Cooklang (.cook) document: YAML
// frontmatter for the metadata, then one section per instruction/ingredient
// group. Ingredients are inlined into the step text where their name appears, so
// the output reads like hand-written Cooklang; ingredients not named in any step
// of their group are emitted as standalone lines within that section.
func marshalCook(ir formats.InterchangeRecipe, w io.Writer) error {
	front, err := marshalFrontmatter(ir)
	if err != nil {
		return err
	}

	var paragraphs []string
	for _, sec := range buildSections(ir) {
		if sec.title != "" {
			paragraphs = append(paragraphs, "== "+sec.title+" ==")
		}
		pool := make([]*placedIngredient, len(sec.ingredients))
		for i, iu := range sec.ingredients {
			pool[i] = &placedIngredient{iu: iu, name: iu.Ingredient.Name}
		}
		for _, step := range sec.steps {
			paragraphs = append(paragraphs, renderStep(step, pool))
		}
		for _, p := range pool {
			if !p.placed {
				paragraphs = append(paragraphs, renderCookIngredient(p.iu))
			}
		}
	}
	if ir.Notes != "" {
		var notes []string
		for _, line := range strings.Split(ir.Notes, "\n") {
			notes = append(notes, "> "+line)
		}
		paragraphs = append(paragraphs, strings.Join(notes, "\n"))
	}

	var b strings.Builder
	if len(front) > 0 {
		b.WriteString("---\n")
		b.Write(front)
		b.WriteString("---\n")
	}
	if len(paragraphs) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(strings.Join(paragraphs, "\n\n"))
		b.WriteString("\n")
	}

	_, err = io.WriteString(w, b.String())
	return err
}

type cookSection struct {
	title       string
	ingredients []ingredients.IngredientUse
	steps       []string
}

type placedIngredient struct {
	iu     ingredients.IngredientUse
	name   string
	placed bool
}

// buildSections groups the interchange ingredients and instructions into ordered
// Cooklang sections keyed by their shared title. Instruction groups define the
// section order; an ingredient group whose title has no matching instruction
// group is appended as its own ingredient-only section.
func buildSections(ir formats.InterchangeRecipe) []*cookSection {
	var order []*cookSection
	byTitle := make(map[string]*cookSection)
	get := func(title string) *cookSection {
		if s, ok := byTitle[title]; ok {
			return s
		}
		s := &cookSection{title: title}
		byTitle[title] = s
		order = append(order, s)
		return s
	}

	for _, tl := range ir.Instructions {
		s := get(tl.Title)
		s.steps = append(s.steps, tl.List...)
	}
	for _, g := range ir.Ingredients {
		s := get(g.Title)
		s.ingredients = append(s.ingredients, g.Items...)
	}
	return order
}

// renderStep inlines any not-yet-placed ingredient from pool whose name appears
// as a whole word in text, replacing the first non-overlapping occurrence with
// its Cooklang markup. Surrounding prose is escaped so it parses back verbatim.
func renderStep(text string, pool []*placedIngredient) string {
	type span struct {
		start, end int
		repl       string
	}
	var spans []span
	overlaps := func(s, e int) bool {
		for _, sp := range spans {
			if s < sp.end && sp.start < e {
				return true
			}
		}
		return false
	}
	for _, p := range pool {
		if p.placed {
			continue
		}
		s, e := findWord(text, p.name)
		if s < 0 || overlaps(s, e) {
			continue
		}
		spans = append(spans, span{s, e, renderCookIngredient(p.iu)})
		p.placed = true
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })

	var b strings.Builder
	last := 0
	for _, sp := range spans {
		b.WriteString(escapeStepText(text[last:sp.start]))
		b.WriteString(sp.repl)
		last = sp.end
	}
	b.WriteString(escapeStepText(text[last:]))
	return b.String()
}

// renderCookIngredient renders one ingredient as Cooklang markup: "@name",
// "@multi word{}", "@name{2}", or "@name{500%g}". Braces are added whenever
// there is a quantity body or the name is more than one word.
func renderCookIngredient(iu ingredients.IngredientUse) string {
	amount, unit := ingredients.CooklangQuantity(iu)
	var body string
	switch {
	case unit != "":
		body = amount + "%" + unit
	case amount != "":
		body = amount
	}
	name := escapeCook(iu.Ingredient.Name)

	var note string
	if iu.Note != "" {
		note = "(" + escapeCook(iu.Note) + ")"
	}
	// A note must hang off a quantity body, so force braces when one is present.
	if body == "" && note == "" && !strings.ContainsAny(iu.Ingredient.Name, " \t") {
		return "@" + name
	}
	return "@" + name + "{" + body + "}" + note
}

type frontmatter struct {
	Title       string   `yaml:"title,omitempty"`
	Description string   `yaml:"description,omitempty"`
	SourceName  string   `yaml:"source.name,omitempty"`
	SourceURL   string   `yaml:"source.url,omitempty"`
	Servings    string   `yaml:"servings,omitempty"`
	PrepTime    string   `yaml:"prep time,omitempty"`
	CookTime    string   `yaml:"cook time,omitempty"`
	TotalTime   string   `yaml:"time,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

func marshalFrontmatter(ir formats.InterchangeRecipe) ([]byte, error) {
	fm := frontmatter{
		Title:       ir.DisplayTitle(),
		Description: ir.Description,
		SourceName:  ir.Source.Name,
		SourceURL:   ir.Source.URI,
		Servings:    ir.Yield,
		PrepTime:    formatCookDuration(ir.PrepTime),
		CookTime:    formatCookDuration(ir.CookTime),
		TotalTime:   formatCookDuration(ir.TotalTime),
		Tags:        ir.Tags,
	}
	empty := fm.Title == "" && fm.Description == "" && fm.SourceName == "" && fm.SourceURL == "" &&
		fm.Servings == "" && fm.PrepTime == "" && fm.CookTime == "" && fm.TotalTime == "" && len(fm.Tags) == 0
	if empty {
		return nil, nil
	}
	return yaml.Marshal(fm)
}

// formatCookDuration renders a duration as a compact Cooklang time ("25m", "1h",
// "1h30m") that MaybeDuration.Parse reads back to the same value.
func formatCookDuration(d *time.Duration) string {
	if d == nil {
		return ""
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh%dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	default:
		return fmt.Sprintf("%dm", m)
	}
}

func findWord(text, word string) (int, int) {
	if word == "" {
		return -1, -1
	}
	lt := strings.ToLower(text)
	lw := strings.ToLower(word)
	if len(lt) != len(text) || len(lw) != len(word) {
		// Lowercasing changed byte lengths (rare non-ASCII); fall back to an
		// exact, case-sensitive search so span indices stay aligned with text.
		lt, lw = text, word
	}
	from := 0
	for {
		i := strings.Index(lt[from:], lw)
		if i < 0 {
			return -1, -1
		}
		start := from + i
		end := start + len(lw)
		if boundaryOK(text, start, end) {
			return start, end
		}
		from = start + 1
	}
}

func boundaryOK(text string, start, end int) bool {
	before := true
	if start > 0 {
		r, _ := utf8.DecodeLastRuneInString(text[:start])
		before = !isWordRune(r)
	}
	after := true
	if end < len(text) {
		r, _ := utf8.DecodeRuneInString(text[end:])
		after = !isWordRune(r)
	}
	return before && after
}

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

var cookEscaper = strings.NewReplacer(
	`\`, `\\`,
	`@`, `\@`,
	`#`, `\#`,
	`~`, `\~`,
	`{`, `\{`,
	`}`, `\}`,
	`%`, `\%`,
)

// escapeCook backslash-escapes the Cooklang inline metacharacters so a literal
// value survives a marshal/parse round-trip.
func escapeCook(s string) string { return cookEscaper.Replace(s) }

// escapeStepText escapes inline metacharacters (via escapeCook) plus the
// line-level markers that would otherwise be read as a comment ("--"), section
// header ("=="), or note (">") when the text is emitted as a step. unescapeCook's
// general "backslash escapes the next character" rule reverses all of these.
func escapeStepText(s string) string {
	s = escapeCook(s)
	s = strings.ReplaceAll(s, "--", `\-\-`)

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		indent := line[:len(line)-len(trimmed)]
		switch {
		case strings.HasPrefix(trimmed, "="):
			lines[i] = indent + `\` + trimmed
		case strings.HasPrefix(trimmed, ">"):
			lines[i] = indent + `\` + trimmed
		}
	}
	return strings.Join(lines, "\n")
}
