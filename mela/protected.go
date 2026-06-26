package mela

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/protected"
)

const protectedExt = ".protectedrecipes"

// ProtectedFormatInfo describes the .protectedrecipes collection: a bundle of Mela
// recipes encrypted so that only someone able to answer questions about the source
// book can decrypt them. The encryption and question machinery live in the root
// `protected` package; this binding supplies and parses the Mela recipe entries.
var ProtectedFormatInfo = &formats.Format{
	Name: "Mela (protected)",
	URL:  "https://github.com/jphastings/recipes/tree/main/protected",
	Features: formats.Features{
		ParseCollection: true,
		WriteCollection: true,
	},
	ExtensionCollection: protectedExt,
	Import:              ImportRecipe,
	NewCollection:       NewProtectedCollection,
	Parse:               ParseProtected,
	Bundle:              formats.BundleByExtension(protectedExt),
}

// ParseProtected decrypts a .protectedrecipes file and streams its Mela recipes.
// It needs ParseOptions.AskOwnership to answer the proof-of-ownership questions;
// without it, protected archives cannot be read.
func ParseProtected(b formats.Bundle, o formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	if o.AskOwnership == nil {
		return nil, nil, errors.New("reading a .protectedrecipes file needs interactive ownership prompts, which aren't available here")
	}
	explain := o.ExplainOwnership
	if explain == nil {
		explain = func(int, int) {}
	}

	filename := b[0]
	// protected.Open prompts (and so decrypts) synchronously here, before any
	// streaming begins.
	entries, closer, err := protected.Open(filename, o.AskOwnership, explain)
	if err != nil {
		return nil, nil, err
	}

	cd := &formats.CollectionDetails{Filename: filename}

	pe := make(chan formats.ParseEvent, 8)
	go func() {
		defer closer.Close()
		defer close(pe)

		pe <- formats.ParseEvent{N: len(entries)}

		for _, e := range entries {
			if strings.HasPrefix(path.Base(e.Name), "._") || !strings.HasSuffix(e.Name, recipeExt) {
				pe <- formats.ParseEvent{I: 1}
				continue
			}

			rr, err := e.Open()
			if err != nil {
				pe <- formats.ParseEvent{Err: err, I: 1}
				continue
			}

			r, err := decodeRecipe(rr, formats.WithoutExt(e.Name))
			rr.Close()
			if err != nil {
				pe <- formats.ParseEvent{Err: fmt.Errorf("couldn't parse recipe within protected archive '%s': %w", e.Name, err), I: 1}
			} else {
				pe <- formats.ParseEvent{Recipe: r, I: 1}
			}
		}
	}()

	return pe, cd, nil
}

var _ formats.CollectionWriter = (*protectedCollectionWriter)(nil)

type protectedCollectionWriter struct {
	filename string
	pw       *protected.Writer
	f        *os.File
}

// NewProtectedCollection creates a new .protectedrecipes collection. The ownership
// questions are generated from the recipes when the collection is closed, so at
// least protected's question-count's worth of recipes must be added first.
func NewProtectedCollection(cd formats.CollectionDetails) (formats.CollectionWriter, error) {
	f, filename, err := formats.OpenCollectionFile(cd, protectedExt)
	if err != nil {
		return nil, err
	}

	return &protectedCollectionWriter{
		filename: filename,
		pw:       protected.NewWriter(f),
		f:        f,
	}, nil
}

func (w *protectedCollectionWriter) Filename() string { return w.filename }

func (w *protectedCollectionWriter) Add(rr formats.Recipe) error {
	r, err := ImportRecipe(rr)
	if err != nil {
		return err
	}
	return w.pw.Add(r)
}

func (w *protectedCollectionWriter) Close() error {
	// pw.Close generates the questions and writes the encrypted archive into f,
	// then the file is closed. A half-written file is removed on any failure.
	err := errors.Join(w.pw.Close(), w.f.Close())
	if err != nil {
		os.Remove(w.filename)
	}
	return err
}
