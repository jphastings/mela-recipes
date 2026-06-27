package schemaorg

import (
	"strconv"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
)

// mapSchemaNode converts a decoded schema.org Recipe node into the interchange
// format, returning any image URLs separately (images are fetched only when the
// caller permits network access).
func mapSchemaNode(node map[string]any) (formats.InterchangeRecipe, []string) {
	ir := formats.NewInterchangeRecipe()

	ir.Title = text(node["name"])
	ir.Description = text(node["description"])
	ir.Yield = text(node["recipeYield"])

	ingredients := node["recipeIngredient"]
	if ingredients == nil {
		ingredients = node["ingredients"] // legacy property name
	}
	if items := textList(ingredients); len(items) > 0 {
		ir.Ingredients = []formats.TitledList{{List: items}}
	}

	ir.Instructions = mapInstructions(node["recipeInstructions"])

	ir.PrepTime = parseISODuration(rawString(node["prepTime"]))
	ir.CookTime = parseISODuration(rawString(node["cookTime"]))
	ir.TotalTime = parseISODuration(rawString(node["totalTime"]))

	ir.Tags = mapTags(node)
	ir.Source = mapSource(node)

	return ir, imageURLs(node["image"])
}

// imageURLs collects the image URL(s) from a schema.org "image" value, which may
// be a URL string, an array of either, or an ImageObject ({ "url": … }).
func imageURLs(v any) []string {
	var out []string
	for _, e := range asSlice(v) {
		switch t := e.(type) {
		case string:
			if u := strings.TrimSpace(t); u != "" {
				out = append(out, u)
			}
		case map[string]any:
			if u := rawString(t["url"]); u != "" {
				out = append(out, u)
			}
		}
	}
	return out
}

// text reduces a JSON value to a single plain-text line, decoding HTML entities
// and stripping tags. It understands strings, numbers, and objects that wrap
// their value in "@value" or "name".
func text(v any) string {
	switch t := v.(type) {
	case string:
		return htmlText(t)
	case float64:
		return formatNumber(t)
	case bool:
		return strconv.FormatBool(t)
	case []any:
		if len(t) > 0 {
			return text(t[0])
		}
	case map[string]any:
		if s := t["@value"]; s != nil {
			return text(s)
		}
		return text(t["name"])
	}
	return ""
}

// rawString coerces a value to a trimmed string without HTML processing — for
// machine values like durations and URLs.
func rawString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return formatNumber(t)
	case []any:
		if len(t) > 0 {
			return rawString(t[0])
		}
	case map[string]any:
		if s := t["@value"]; s != nil {
			return rawString(s)
		}
		if s := t["@id"]; s != nil {
			return rawString(s)
		}
		return rawString(t["url"])
	}
	return ""
}

func formatNumber(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// asSlice normalises a single-or-array value to a slice.
func asSlice(v any) []any {
	switch t := v.(type) {
	case nil:
		return nil
	case []any:
		return t
	default:
		return []any{t}
	}
}

// textList flattens a value into a list of cleaned text items, dropping empties.
func textList(v any) []string {
	var out []string
	for _, e := range asSlice(v) {
		if s := text(e); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// mapInstructions handles every shape recipeInstructions takes: a single string,
// an array of strings or HowToStep objects (one untitled section), or an array of
// HowToSection objects (one TitledList each).
func mapInstructions(v any) []formats.TitledList {
	switch t := v.(type) {
	case string:
		if lines := htmlLines(t); len(lines) > 0 {
			return []formats.TitledList{{List: lines}}
		}
		return nil
	case map[string]any:
		return mapInstructions([]any{t})
	case []any:
		var sections []formats.TitledList
		var loose []string
		for _, e := range t {
			if m, ok := e.(map[string]any); ok && isType(m["@type"], "HowToSection") {
				if steps := stepTexts(m["itemListElement"]); len(steps) > 0 {
					sections = append(sections, formats.TitledList{Title: text(m["name"]), List: steps})
				}
				continue
			}
			loose = append(loose, stepTexts(e)...)
		}
		if len(loose) > 0 {
			sections = append([]formats.TitledList{{List: loose}}, sections...)
		}
		return sections
	}
	return nil
}

// stepTexts extracts step text from a value that may be a string, a HowToStep /
// HowToDirection object, or a nested list of them.
func stepTexts(v any) []string {
	switch t := v.(type) {
	case string:
		if s := htmlText(t); s != "" {
			return []string{s}
		}
		return nil
	case []any:
		var out []string
		for _, e := range t {
			out = append(out, stepTexts(e)...)
		}
		return out
	case map[string]any:
		if nested := t["itemListElement"]; nested != nil {
			if out := stepTexts(nested); len(out) > 0 {
				return out
			}
		}
		s := text(t["text"])
		if s == "" {
			s = text(t["name"])
		}
		if s == "" {
			return nil
		}
		return []string{s}
	}
	return nil
}

// mapTags merges recipeCategory, recipeCuisine and keywords (any of which may be
// a string, an array, or a comma-separated string) into a deduplicated tag list.
func mapTags(node map[string]any) []string {
	var tags []string
	add := func(v any) {
		for _, e := range asSlice(v) {
			for _, part := range strings.Split(text(e), ",") {
				if p := strings.TrimSpace(part); p != "" {
					tags = append(tags, p)
				}
			}
		}
	}
	add(node["recipeCategory"])
	add(node["recipeCuisine"])
	add(node["keywords"])
	return dedupe(tags)
}

func dedupe(in []string) []string {
	seen := make(map[string]bool, len(in))
	var out []string
	for _, s := range in {
		key := strings.ToLower(s)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, s)
	}
	return out
}

// mapSource records where the recipe came from: the canonical page URL, and the
// author's name when present.
func mapSource(node map[string]any) formats.Source {
	uri := rawString(node["url"])
	if uri == "" {
		uri = rawString(node["@id"])
	}
	return formats.Source{
		Name: text(node["author"]),
		URI:  uri,
	}
}
