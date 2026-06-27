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
// When opts.AllowNetwork is set, the recipe's images are fetched and embedded.
func Parse(b formats.Bundle, opts formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	filename := b[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	pe := make(chan formats.ParseEvent, 1)
	go func() {
		defer close(pe)
		ir, err := extractRecipe(data, strings.ToLower(path.Ext(filename)), opts.AllowNetwork)
		if err != nil {
			pe <- formats.ParseEvent{Err: fmt.Errorf("%s: %w", filename, err), I: 1, N: 1}
			return
		}
		pe <- formats.ParseEvent{Recipe: &Recipe{ir: ir}, I: 1, N: 1}
	}()

	return pe, nil, nil
}

// extractRecipe maps the recipe from the file's structured data and, when network
// access is permitted, fetches its images.
func extractRecipe(data []byte, ext string, allowNetwork bool) (formats.InterchangeRecipe, error) {
	ir, imageURLs, err := parseStructured(data, ext)
	if err != nil {
		return formats.InterchangeRecipe{}, err
	}
	if allowNetwork {
		if imgs := fetchImages(imageURLs); len(imgs) > 0 {
			ir.Images = imgs
		}
	}
	return validated(ir)
}

// parseStructured maps the highest-priority structured-recipe data found in the
// file — JSON-LD first, then microdata, then the h-recipe microformat — returning
// the recipe and any image URLs it references.
func parseStructured(data []byte, ext string) (formats.InterchangeRecipe, []string, error) {
	if ext == ".json" {
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return formats.InterchangeRecipe{}, nil, fmt.Errorf("invalid JSON: %w", err)
		}
		node, ok := findRecipeNode(v)
		if !ok {
			return formats.InterchangeRecipe{}, nil, errNoRecipe
		}
		ir, urls := mapSchemaNode(node)
		return ir, urls, nil
	}

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return formats.InterchangeRecipe{}, nil, fmt.Errorf("invalid HTML: %w", err)
	}
	if node, ok := jsonLDRecipe(doc); ok {
		ir, urls := mapSchemaNode(node)
		return ir, urls, nil
	}
	if ir, urls, ok := microdataRecipe(doc); ok {
		return ir, urls, nil
	}
	if ir, urls, ok := hrecipeRecipe(doc); ok {
		return ir, urls, nil
	}
	return formats.InterchangeRecipe{}, nil, errNoRecipe
}

func validated(ir formats.InterchangeRecipe) (formats.InterchangeRecipe, error) {
	if errs := ir.Validate(); len(errs) > 0 {
		return ir, fmt.Errorf("incomplete recipe: %w", errors.Join(errs...))
	}
	return ir, nil
}
