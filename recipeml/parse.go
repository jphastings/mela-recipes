package recipeml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"

	"github.com/jphastings/recipes/internal/formats"
	"golang.org/x/net/html/charset"
)

// document mirrors a RecipeML file: a <recipeml> root holding one or more
// <recipe> elements.
type document struct {
	XMLName xml.Name `xml:"recipeml"`
	Recipes []recipe `xml:"recipe"`
}

// Parse reads a single `.xml` file and streams one ParseEvent per <recipe> it
// contains. A RecipeML document is a flat list of recipes, not a named app
// collection, so no CollectionDetails are returned. XML that isn't RecipeML
// yields a single non-fatal error event.
func Parse(b formats.Bundle, _ formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	filename := b[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	doc, err := decode(data)
	if err != nil {
		return errorEvent(fmt.Errorf("%s: %w", filename, err)), nil, nil
	}
	if doc.XMLName.Local != "recipeml" {
		return errorEvent(fmt.Errorf("%s: not a RecipeML document (root is <%s>)", filename, doc.XMLName.Local)), nil, nil
	}

	pe := make(chan formats.ParseEvent, len(doc.Recipes)+1)
	go func() {
		defer close(pe)
		pe <- formats.ParseEvent{N: len(doc.Recipes)}

		for _, xr := range doc.Recipes {
			ir, err := toInterchange(xr)
			if err == nil {
				if errs := ir.Validate(); len(errs) > 0 {
					err = fmt.Errorf("incomplete RecipeML recipe %q: %w", ir.Title, errors.Join(errs...))
				}
			}
			if err != nil {
				pe <- formats.ParseEvent{Err: fmt.Errorf("%s: %w", filename, err), I: 1}
				continue
			}
			pe <- formats.ParseEvent{Recipe: &Recipe{ir: ir}, I: 1}
		}
	}()

	return pe, nil, nil
}

// decode parses the document leniently. CharsetReader honours a declared encoding
// (RecipeML archives are often ISO-8859-1); external DTDs are never fetched, so
// there is no XXE exposure.
func decode(data []byte) (document, error) {
	var doc document
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity
	dec.CharsetReader = charset.NewReaderLabel
	if err := dec.Decode(&doc); err != nil {
		return document{}, err
	}
	return doc, nil
}

func errorEvent(err error) <-chan formats.ParseEvent {
	pe := make(chan formats.ParseEvent, 1)
	pe <- formats.ParseEvent{Err: err}
	close(pe)
	return pe
}
