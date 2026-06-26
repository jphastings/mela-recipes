package cooklang

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
	"github.com/jphastings/recipes/utils"
	"gopkg.in/yaml.v3"
)

func Parse(b formats.Bundle, _ formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	pe := make(chan formats.ParseEvent)

	go func(pe chan formats.ParseEvent, b formats.Bundle) {
		pe <- formats.ParseEvent{N: 1}
		r, err := ParseRecipe(b)
		pe <- formats.ParseEvent{Recipe: r, Err: err, I: 1}
		close(pe)
	}(pe, b)

	return pe, nil, nil
}

func ParseRecipe(b formats.Bundle) (formats.Recipe, error) {
	cookFile, imageFiles := b[0], b[1:]

	f, err := os.Open(cookFile)
	if err != nil {
		return nil, err
	}
	ir, err := parseCook(f)
	closeErr := f.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}

	for _, imageFile := range imageFiles {
		img, err := readImage(imageFile)
		if err != nil {
			return nil, err
		}
		ir.Images = append(ir.Images, img)
	}

	return &Recipe{filename: formats.WithoutExt(filepath.Base(cookFile)), ir: ir}, nil
}

func readImage(path string) (utils.B64Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return utils.FromFile(f)
}

// parseCook reads a Cooklang document into an interchange recipe. It is the
// inverse of marshalCook: frontmatter becomes the scalar metadata, each
// "== Title ==" section becomes an ingredient and/or instruction group of that
// title, inline @/#/~ markup is extracted (cookware and timers are folded into
// the step text), and "> " lines become the notes.
func parseCook(r io.Reader) (formats.InterchangeRecipe, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return formats.InterchangeRecipe{}, err
	}
	s := strings.ReplaceAll(string(raw), "\r\n", "\n")

	metaStr, body := splitFrontmatter(s)
	ir := formats.NewInterchangeRecipe()
	if metaStr != "" {
		var meta map[string]any
		if err := yaml.Unmarshal([]byte(metaStr), &meta); err == nil {
			applyFrontmatter(&ir, meta)
		}
	}

	body = stripComments(body)

	type rawSection struct {
		title string
		lines []string
	}
	cur := &rawSection{}
	secs := []*rawSection{cur}
	var notes []string
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		switch {
		case strings.HasPrefix(trimmed, "="):
			cur = &rawSection{title: strings.TrimSpace(strings.Trim(trimmed, "= \t"))}
			secs = append(secs, cur)
		case strings.HasPrefix(trimmed, ">>"):
			// Legacy ">>" metadata is not emitted by this writer; ignore on read.
		case strings.HasPrefix(trimmed, ">"):
			notes = append(notes, strings.TrimSpace(strings.TrimPrefix(trimmed, ">")))
		default:
			cur.lines = append(cur.lines, line)
		}
	}

	for _, sec := range secs {
		var ingLines, instrLines []string
		order := 0
		for _, block := range splitSteps(sec.lines) {
			instr, ings := parseStep(block, &order)
			ingLines = append(ingLines, ings...)
			if instr != "" {
				instrLines = append(instrLines, instr)
			}
		}
		if len(ingLines) > 0 {
			ir.Ingredients = append(ir.Ingredients, formats.TitledList{Title: sec.title, List: ingLines})
		}
		if len(instrLines) > 0 {
			ir.Instructions = append(ir.Instructions, formats.TitledList{Title: sec.title, List: instrLines})
		}
	}

	ir.Notes = strings.Join(notes, "\n")
	return ir, nil
}

func splitFrontmatter(s string) (meta string, body string) {
	s = strings.TrimPrefix(s, "\ufeff")
	if !strings.HasPrefix(s, "---\n") {
		return "", s
	}
	lines := strings.Split(s[len("---\n"):], "\n")
	for i, l := range lines {
		if t := strings.TrimRight(l, " \t"); t == "---" || t == "..." {
			return strings.Join(lines[:i], "\n"), strings.Join(lines[i+1:], "\n")
		}
	}
	return "", s
}

func applyFrontmatter(ir *formats.InterchangeRecipe, meta map[string]any) {
	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := meta[k]; ok {
				if s := scalarString(v); s != "" {
					return s
				}
			}
		}
		return ""
	}

	ir.Title = get("title")
	ir.Description = get("description", "introduction")
	ir.Source.Name = get("source.name", "author")
	ir.Source.URI = get("source.url", "source.uri")
	if ir.Source.Name == "" && ir.Source.URI == "" {
		if src := get("source"); src != "" {
			if looksLikeURI(src) {
				ir.Source.URI = src
			} else {
				ir.Source.Name = src
			}
		}
	}
	ir.Yield = get("servings", "serves", "yield")
	if d, err := formats.MaybeDuration(get("prep time", "time.prep")).Parse(); err == nil {
		ir.PrepTime = d
	}
	if d, err := formats.MaybeDuration(get("cook time", "time.cook")).Parse(); err == nil {
		ir.CookTime = d
	}
	if d, err := formats.MaybeDuration(get("time", "time required", "duration", "total time")).Parse(); err == nil {
		ir.TotalTime = d
	}
	ir.Tags = stringList(meta["tags"])
}

func scalarString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprint(x)
	}
}

func stringList(v any) []string {
	switch x := v.(type) {
	case []any:
		var out []string
		for _, e := range x {
			if s := scalarString(e); s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if x == "" {
			return nil
		}
		return []string{x}
	default:
		return nil
	}
}

func looksLikeURI(s string) bool {
	return strings.Contains(s, "://") || strings.HasPrefix(s, "urn:")
}

// stripComments removes Cooklang line ("-- …") and block ("[- … -]") comments,
// while preserving backslash escape sequences so escaped metacharacters are not
// mistaken for comment markers.
func stripComments(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '\\' {
			b.WriteByte('\\')
			i++
			if i < len(s) {
				_, sz := utf8.DecodeRuneInString(s[i:])
				b.WriteString(s[i : i+sz])
				i += sz
			}
			continue
		}
		if strings.HasPrefix(s[i:], "[-") {
			if end := strings.Index(s[i+2:], "-]"); end >= 0 {
				i += 2 + end + 2
			} else {
				i = len(s)
			}
			continue
		}
		if strings.HasPrefix(s[i:], "--") {
			if nl := strings.IndexByte(s[i:], '\n'); nl >= 0 {
				i += nl
			} else {
				i = len(s)
			}
			continue
		}
		_, sz := utf8.DecodeRuneInString(s[i:])
		b.WriteString(s[i : i+sz])
		i += sz
	}
	return b.String()
}

// splitSteps groups a section's lines into steps, separated by blank lines.
// Lines within a step keep their order and single newlines.
func splitSteps(lines []string) []string {
	var blocks []string
	var cur []string
	flush := func() {
		if len(cur) > 0 {
			blocks = append(blocks, strings.Join(cur, "\n"))
			cur = nil
		}
	}
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			flush()
		} else {
			cur = append(cur, strings.TrimRight(l, " \t"))
		}
	}
	flush()
	return blocks
}

type token struct {
	sigil byte
	name  string
	qty   string
	unit  string
}

// parseStep extracts inline markup from a single step. Ingredients become
// canonical interchange strings (via the ingredients formatter); cookware and
// timers are rendered to their display text and folded into the prose. A step
// that is nothing but ingredient markup is treated as a standalone ingredient
// line and contributes no instruction text.
func parseStep(text string, order *int) (instruction string, ingLines []string) {
	var display, connective strings.Builder
	ingCount := 0

	for i := 0; i < len(text); {
		c := text[i]
		if c == '\\' && i+1 < len(text) {
			_, sz := utf8.DecodeRuneInString(text[i+1:])
			ch := text[i+1 : i+1+sz]
			display.WriteString(ch)
			connective.WriteString(ch)
			i += 1 + sz
			continue
		}
		if c == '@' || c == '#' || c == '~' {
			tok, next := readToken(text, i)
			if next > i {
				switch tok.sigil {
				case '@':
					if iu, err := ingredients.FromCooklang(tok.qty, tok.unit, tok.name, *order); err == nil {
						ingLines = append(ingLines, ingredients.FormatIngredientUse(iu))
						*order++
						ingCount++
					}
					display.WriteString(tok.name)
				case '#':
					display.WriteString(tok.name)
				case '~':
					display.WriteString(timerDisplay(tok))
				}
				i = next
				continue
			}
		}
		_, sz := utf8.DecodeRuneInString(text[i:])
		display.WriteString(text[i : i+sz])
		connective.WriteString(text[i : i+sz])
		i += sz
	}

	if ingCount > 0 && !hasLetterOrDigit(connective.String()) {
		return "", ingLines
	}
	return strings.TrimSpace(display.String()), ingLines
}

// readToken parses a single @/#/~ reference beginning at text[i]. A name is read
// up to a delimiting "{…}" quantity (allowing spaces, so multiword names work),
// or as a single bareword when no brace follows.
func readToken(text string, i int) (token, int) {
	tok := token{sigil: text[i]}
	j := i + 1

	bracePos := -1
	for k := j; k < len(text); {
		c := text[k]
		if c == '\\' {
			k += 2
			continue
		}
		if c == '{' {
			bracePos = k
			break
		}
		if c == '@' || c == '#' || c == '~' || c == '\n' {
			break
		}
		k++
	}

	var next int
	if bracePos >= 0 {
		if end := strings.IndexByte(text[bracePos+1:], '}'); end >= 0 {
			tok.name = unescapeCook(text[j:bracePos])
			qty := text[bracePos+1 : bracePos+1+end]
			next = bracePos + 1 + end + 1
			if parts := strings.SplitN(qty, "%", 2); len(parts) == 2 {
				tok.qty = strings.TrimSpace(unescapeCook(parts[0]))
				tok.unit = strings.TrimSpace(unescapeCook(parts[1]))
			} else {
				tok.qty = strings.TrimSpace(unescapeCook(parts[0]))
			}
			// Skip an optional "(preparation)" suffix; it is folded away.
			if next < len(text) && text[next] == '(' {
				if pend := strings.IndexByte(text[next+1:], ')'); pend >= 0 {
					next = next + 1 + pend + 1
				}
			}
			return tok, next
		}
	}

	// Bareword: a run of word characters (escapes included).
	e := j
	for e < len(text) {
		if text[e] == '\\' {
			e++
			if e < len(text) {
				_, sz := utf8.DecodeRuneInString(text[e:])
				e += sz
			}
			continue
		}
		r, sz := utf8.DecodeRuneInString(text[e:])
		if !isWordRune(r) {
			break
		}
		e += sz
	}
	tok.name = unescapeCook(text[j:e])
	return tok, e
}

func timerDisplay(t token) string {
	switch {
	case t.qty != "" && t.unit != "":
		return t.qty + " " + t.unit
	case t.qty != "":
		return t.qty
	default:
		return t.name
	}
}

func hasLetterOrDigit(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// unescapeCook reverses escapeCook/escapeStepText: a backslash escapes the
// following character.
func unescapeCook(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '\\' && i+1 < len(s) {
			_, sz := utf8.DecodeRuneInString(s[i+1:])
			b.WriteString(s[i+1 : i+1+sz])
			i += 1 + sz
			continue
		}
		_, sz := utf8.DecodeRuneInString(s[i:])
		b.WriteString(s[i : i+sz])
		i += sz
	}
	return b.String()
}
