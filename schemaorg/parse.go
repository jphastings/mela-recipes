package schemaorg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"golang.org/x/net/html"
)

var errNoRecipe = errors.New("no schema.org Recipe found (looked for JSON-LD, microdata, and h-recipe)")

// Parse reads a single .html/.htm/.json file and emits the recipe it describes.
func Parse(b formats.Bundle, _ formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	filename := b[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	pe := make(chan formats.ParseEvent, 1)
	go func() {
		defer close(pe)
		ir, err := extractRecipe(data, strings.ToLower(path.Ext(filename)))
		if err != nil {
			pe <- formats.ParseEvent{Err: fmt.Errorf("%s: %w", filename, err), I: 1, N: 1}
			return
		}
		pe <- formats.ParseEvent{Recipe: &Recipe{ir: ir}, I: 1, N: 1}
	}()

	return pe, nil, nil
}

// extractRecipe maps the highest-priority structured-recipe data found in the
// file: JSON-LD first, then microdata, then the h-recipe microformat.
func extractRecipe(data []byte, ext string) (formats.InterchangeRecipe, error) {
	if ext == ".json" {
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return formats.InterchangeRecipe{}, fmt.Errorf("invalid JSON: %w", err)
		}
		node, ok := findRecipeNode(v)
		if !ok {
			return formats.InterchangeRecipe{}, errNoRecipe
		}
		return validated(mapSchemaNode(node))
	}

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return formats.InterchangeRecipe{}, fmt.Errorf("invalid HTML: %w", err)
	}
	if node, ok := jsonLDRecipe(doc); ok {
		return validated(mapSchemaNode(node))
	}
	if ir, ok := microdataRecipe(doc); ok {
		return validated(ir)
	}
	if ir, ok := hrecipeRecipe(doc); ok {
		return validated(ir)
	}
	return formats.InterchangeRecipe{}, errNoRecipe
}

func validated(ir formats.InterchangeRecipe) (formats.InterchangeRecipe, error) {
	if errs := ir.Validate(); len(errs) > 0 {
		return ir, fmt.Errorf("incomplete recipe: %w", errors.Join(errs...))
	}
	return ir, nil
}
