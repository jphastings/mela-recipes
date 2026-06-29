package recipes

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/jphastings/recipes/cooklang"
	"github.com/jphastings/recipes/crouton"
	"github.com/jphastings/recipes/epub"
	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/mealmaster"
	"github.com/jphastings/recipes/mela"
	"github.com/jphastings/recipes/paprika"
	"github.com/jphastings/recipes/recipemd"
	"github.com/jphastings/recipes/recipeml"
	"github.com/jphastings/recipes/schemaorg"
)

func AvailableFormats() []*formats.Format {
	return []*formats.Format{
		mela.FormatInfo,
		mela.ProtectedFormatInfo,
		crouton.FormatInfo,
		paprika.FormatInfo,
		epub.FormatInfo,
		cooklang.FormatInfo,
		recipemd.FormatInfo,
		schemaorg.FormatInfo,
		mealmaster.FormatInfo,
		recipeml.FormatInfo,
	}
}

type rollupProgress struct {
	totals map[string]int
	mu     *sync.Mutex
}

func newRollupProgress() rollupProgress {
	return rollupProgress{
		mu:     &sync.Mutex{},
		totals: make(map[string]int),
	}
}

// Adds the totals together
func (rp rollupProgress) Add(idx string, thisN int) int {
	rp.mu.Lock()
	rp.totals[idx] = thisN

	rollupN := 0
	for _, t := range rp.totals {
		rollupN += t
	}

	rp.mu.Unlock()
	return rollupN
}

// Attempts to find a suitable format & parser for all the recipe files given. All found recipes are returned in the first argument, the second argument holds the details of the collection if the input files represent *exactly and only one* collection.
// If no collections are represented, or the files represent a collection *and* other recipes or collections, then this second return value will be nil.
func ParseAll(files []string, o formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	moreThanOneCollection := false
	var cd *formats.CollectionDetails
	pe := make(chan formats.ParseEvent)
	rp := newRollupProgress()
	var wg sync.WaitGroup

	for _, f := range AvailableFormats() {
		bundles, unused := f.Bundle(files)
		files = unused

		for ib, b := range bundles {
			idx := fmt.Sprintf("%s-%d", f.Name, ib)

			bpe, bcd, err := f.Parse(b, o)
			if err != nil {
				return nil, nil, err
			}

			// If only one set of collection details are present then pass on those, otherwise treat all collections as bunch of recipes
			if bcd != nil && !moreThanOneCollection {
				if cd != nil {
					cd = nil
					moreThanOneCollection = true
				} else {
					cd = bcd
				}
			}

			wg.Add(1)
			go func(pe chan<- formats.ParseEvent, bpe <-chan formats.ParseEvent, totals rollupProgress, idx string, wg *sync.WaitGroup) {
				for e := range bpe {
					n := totals.Add(idx, e.N)

					// Emit with the new rolled-up progress counter
					pe <- formats.ParseEvent{
						Recipe: e.Recipe,
						Err:    e.Err,
						I:      e.I,
						N:      n,
					}
				}
				wg.Done()
			}(pe, bpe, rp, idx, &wg)
		}
	}

	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(pe)
	}(&wg)

	return pe, cd, nil
}

type AsType string

const (
	AsTypeAny        AsType = "any"
	AsTypeRecipe     AsType = "recipe"
	AsTypeCollection AsType = "collection"
)

func ParseDestination(to string) (overrideFilename string, asType AsType, format *formats.Format) {
	ext := path.Ext(to)
	if ext != "" && ext != to {
		overrideFilename = strings.TrimSuffix(to, ext)
	}

	for _, f := range AvailableFormats() {
		if f.Extension != "" && f.Extension == ext {
			return overrideFilename, AsTypeRecipe, f
		}
		if f.ExtensionCollection != "" && f.ExtensionCollection == ext {
			return overrideFilename, AsTypeCollection, f
		}
		if strings.EqualFold(to, f.Name) {
			return "", AsTypeAny, f
		}
	}

	return "", AsTypeAny, nil
}
