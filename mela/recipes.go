package mela

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/yeka/zip"
)

var _ formats.CollectionWriter = (*collectionWriter)(nil)

type collectionWriter struct {
	filename string
	z        *zip.Writer
	close    func() error
}

// ParseRecipe parses a known .melarecipes collection file into a RecipeCollection compatible struct.
func ParseRecipesFile(filename string) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	zr, err := zip.OpenReader(filename)
	if err != nil {
		return nil, nil, err
	}
	// Don't defer close; it needs to remain open as long as the goroutine is live

	// TODO: Can we grab any book details from the zip?
	cd := &formats.CollectionDetails{
		Filename: filename,
	}

	pe := make(chan formats.ParseEvent, 8)
	go func(pe chan formats.ParseEvent, zr *zip.ReadCloser) {
		n := len(zr.File)
		pe <- formats.ParseEvent{N: n}

		for _, zf := range zr.File {
			if strings.HasPrefix(path.Base(zf.Name), "._") || !strings.HasSuffix(zf.Name, recipeExt) {
				pe <- formats.ParseEvent{I: 1}
				continue
			}

			rr, err := zf.Open()
			if err != nil {
				pe <- formats.ParseEvent{Err: err, I: 1}
				continue
			}
			defer rr.Close()

			r, err := ParseRecipeStream(rr)
			if err != nil {
				pe <- formats.ParseEvent{
					Err: fmt.Errorf("couldn't parse recipe within zip '%s': %w", zf.Name, err),
					I:   1,
				}
			} else {
				r.filename = withoutExt(zf.Name)
				pe <- formats.ParseEvent{Recipe: r, I: 1}
			}
		}
		zr.Close()
		close(pe)
	}(pe, zr)

	return pe, cd, nil
}

func NewCollection(cd formats.CollectionDetails) (formats.CollectionWriter, error) {
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
	close := func() error {
		return errors.Join(z.Close(), f.Close())
	}

	return &collectionWriter{filename: filename, z: z, close: close}, nil
}

func (cw *collectionWriter) Filename() string { return cw.filename }

func (cw *collectionWriter) Close() error { return cw.close() }

func (cw *collectionWriter) Add(rr formats.Recipe) error {
	ir, err := FormatInfo.Import(rr)
	if err != nil {
		return err
	}
	r := ir.(*Recipe)

	w, err := cw.z.Create(r.Filename())
	if err != nil {
		return fmt.Errorf("unable to create recipe file in zip: %w", err)
	}

	if err := r.Marshal(w); err != nil {
		return fmt.Errorf("unable to encode recipe JSON into zip: %w", err)
	}

	return nil
}
