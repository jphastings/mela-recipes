package formats

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/yeka/zip"
)

// RecipeDecodeFunc decodes a single recipe of a particular format from a stream.
// filename is the recipe's name without its extension, to be stored on the
// returned recipe so it can round-trip. Implementations own any per-format
// container handling (eg. gzip) before decoding the recipe body.
type RecipeDecodeFunc func(r io.Reader, filename string) (Recipe, error)

// WithoutExt returns name with its final extension removed.
func WithoutExt(name string) string {
	return strings.TrimSuffix(name, path.Ext(name))
}

// ParseRecipeFile opens a single-recipe file and streams exactly one
// ParseEvent, delegating the decode (and any decompression) to decode.
func ParseRecipeFile(filename string, decode RecipeDecodeFunc) (<-chan ParseEvent, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	// Don't defer close; it needs to remain open as long as the goroutine is live

	pe := make(chan ParseEvent)
	go func() {
		r, err := decode(f, WithoutExt(filename))
		f.Close()

		pe <- ParseEvent{Recipe: r, Err: err, I: 1, N: 1}
		close(pe)
	}()

	return pe, nil
}

// ParseZipCollection streams the recipes out of a zip archive (eg. Mela's
// .melarecipes, Paprika's .paprikarecipes). Entries whose names don't end in
// recipeExt, and macOS "._" resource forks, are skipped. Each kept entry is
// handed to decode.
func ParseZipCollection(filename, recipeExt string, decode RecipeDecodeFunc) (<-chan ParseEvent, *CollectionDetails, error) {
	zr, err := zip.OpenReader(filename)
	if err != nil {
		return nil, nil, err
	}
	// Don't defer close; it needs to remain open as long as the goroutine is live

	cd := &CollectionDetails{Filename: filename}

	pe := make(chan ParseEvent, 8)
	go func() {
		pe <- ParseEvent{N: len(zr.File)}

		for _, zf := range zr.File {
			if strings.HasPrefix(path.Base(zf.Name), "._") || !strings.HasSuffix(zf.Name, recipeExt) {
				pe <- ParseEvent{I: 1}
				continue
			}

			rr, err := zf.Open()
			if err != nil {
				pe <- ParseEvent{Err: err, I: 1}
				continue
			}

			r, err := decode(rr, WithoutExt(zf.Name))
			rr.Close()
			if err != nil {
				pe <- ParseEvent{Err: fmt.Errorf("couldn't parse recipe within zip '%s': %w", zf.Name, err), I: 1}
			} else {
				pe <- ParseEvent{Recipe: r, I: 1}
			}
		}

		zr.Close()
		close(pe)
	}()

	return pe, cd, nil
}

var _ CollectionWriter = (*zipCollectionWriter)(nil)

type zipCollectionWriter struct {
	filename string
	z        *zip.Writer
	importFn func(Recipe) (Recipe, error)
	close    func() error
}

// NewZipCollection opens a new zip-backed recipe collection. Recipes added to it
// are first converted with importFn, then written via their own Marshal, so the
// per-format container encoding (eg. gzip) lives entirely in the recipe type.
func NewZipCollection(cd CollectionDetails, collectionExt string, importFn func(Recipe) (Recipe, error)) (CollectionWriter, error) {
	filename := cd.Filename + collectionExt
	flags := os.O_CREATE
	if cd.OverwriteExisting {
		flags |= os.O_TRUNC | os.O_RDWR
	} else {
		flags |= os.O_EXCL | os.O_WRONLY
	}

	f, err := os.OpenFile(filename, flags, 0644)
	if err != nil {
		return nil, err
	}
	z := zip.NewWriter(f)
	closeFn := func() error {
		return errors.Join(z.Close(), f.Close())
	}

	return &zipCollectionWriter{filename: filename, z: z, importFn: importFn, close: closeFn}, nil
}

func (cw *zipCollectionWriter) Filename() string { return cw.filename }

func (cw *zipCollectionWriter) Close() error { return cw.close() }

func (cw *zipCollectionWriter) Add(rr Recipe) error {
	r, err := cw.importFn(rr)
	if err != nil {
		return err
	}

	w, err := cw.z.Create(r.Filename())
	if err != nil {
		return fmt.Errorf("unable to create recipe file in zip: %w", err)
	}

	if err := r.Marshal(w); err != nil {
		return fmt.Errorf("unable to encode recipe into zip: %w", err)
	}

	return nil
}
