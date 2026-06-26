package epub

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/jphastings/recipes/epub/induce"
	"github.com/jphastings/recipes/epub/induce/modellabel"
	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/utils"
	"github.com/pirmd/epub"
)

// modelRefiner is a process-wide, lazily-loaded model labeller. The embedding
// model is loaded only the first time a book can't be certified structurally,
// so self-certifying books pay nothing, and a load failure (e.g. offline)
// degrades to structural-only extraction.
var modelRefiner = &lazyLabeler{}

type lazyLabeler struct {
	once sync.Once
	lab  induce.Labeler
}

func (z *lazyLabeler) Label(units []induce.Unit, unit induce.UnitSpec) map[induce.Role]induce.FieldSpec {
	z.once.Do(func() {
		if l, err := modellabel.New(context.Background(), modelCacheDir()); err == nil {
			z.lab = l
		}
	})
	if z.lab == nil {
		return map[induce.Role]induce.FieldSpec{}
	}
	return z.lab.Label(units, unit)
}

func modelCacheDir() string {
	if d := os.Getenv("RECIPES_MODEL_DIR"); d != "" {
		return d
	}
	if d, err := os.UserCacheDir(); err == nil {
		return filepath.Join(d, "recipes", "models")
	}
	return "./models"
}

const (
	minPhotoDim = 300 // below this on either side it's an icon/ornament, not a recipe photo
	maxPhotos   = 3   // attach at most this many photos per recipe
)

// photoFilter decides whether an image is a recipe photo. The byte-size floor is
// learned from the book itself: flat diagrams and logos compress far smaller than
// photographs, so a fraction of the typical photo's size cleanly separates them.
type photoFilter struct {
	minBytes int
	minDim   int
}

func isImageItem(mediaType, href string) bool {
	if strings.HasPrefix(mediaType, "image/") {
		return true
	}
	l := strings.ToLower(href)
	return strings.HasSuffix(l, ".jpg") || strings.HasSuffix(l, ".jpeg") || strings.HasSuffix(l, ".png")
}

// buildPhotoFilter scans every image once (header + stored size only) to learn
// the book's typical photo size, then sets the floor to a quarter of it.
func buildPhotoFilter(e *epub.Epub) photoFilter {
	filt := photoFilter{minDim: minPhotoDim}
	pkg, err := e.Package()
	if err != nil {
		return filt
	}
	var sizes []int
	for _, it := range pkg.Manifest.Items {
		if !isImageItem(it.MediaType, it.Href) {
			continue
		}
		f, err := e.OpenItem(it.Href)
		if err != nil {
			continue
		}
		var size int
		if fi, err := f.Stat(); err == nil {
			size = int(fi.Size())
		}
		cfg, format, derr := safeDecodeConfig(f)
		f.Close()
		if derr != nil || (format != "jpeg" && format != "png") {
			continue
		}
		if size > 0 && cfg.Width >= minPhotoDim && cfg.Height >= minPhotoDim {
			sizes = append(sizes, size)
		}
	}
	if len(sizes) >= 4 {
		sort.Ints(sizes)
		upper := sizes[len(sizes)/2:] // larger half ≈ the photographs
		typical := upper[len(upper)/2]
		filt.minBytes = typical / 4
	}
	return filt
}

// Parse turns an ePub cookbook into a stream of recipes. It induces the book's
// own recipe layout (no per-book configuration, no LLM) and extracts every
// recipe verbatim, so the result is exactly what is printed in the book.
func Parse(b formats.Bundle, o formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	if len(b) != 1 {
		return nil, nil, fmt.Errorf("ePub bundles must contain exactly one ePub filename")
	}
	filename := b[0]
	if path.Ext(filename) != collectionExt {
		return nil, nil, fmt.Errorf("doesn't appear to be an ePub file")
	}

	e, err := openEpub(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't open ePub: %w", err)
	}

	cd := &formats.CollectionDetails{
		Filename: strings.TrimSuffix(filename, collectionExt),
	}
	// Use the ePub's title, if it's in the standard place
	if i, err := e.Information(); err == nil && len(i.Title) > 0 {
		cd.Name = i.Title[0]
	}

	pe := make(chan formats.ParseEvent)
	go func() {
		defer e.Close()
		defer close(pe)
		// Backstop: a codec (eg. jpegli) trap surfaces as a panic. Turn any that
		// escape the per-image guards into an error rather than crashing the CLI.
		defer func() {
			if r := recover(); r != nil {
				pe <- formats.ParseEvent{Err: fmt.Errorf("ePub extraction failed unexpectedly: %v", r)}
			}
		}()
		extractRecipes(e, pe)
	}()

	return pe, cd, nil
}

// openEpub wraps epub.Open, which panics on some malformed archives instead of
// returning an error.
func openEpub(filename string) (e *epub.Epub, err error) {
	defer func() {
		if r := recover(); r != nil {
			e, err = nil, fmt.Errorf("malformed or unsupported ePub archive (%v)", r)
		}
	}()
	return epub.Open(filename)
}

var (
	devNull     *os.File
	devNullOnce sync.Once
	stderrMu    sync.Mutex
)

// withQuietStderr runs fn with os.Stderr pointed at the null device. The jpegli
// decoder writes chatty C-level diagnostics ("Skipped 1 bytes before marker…")
// straight to stderr on malformed images, with no API to disable them; this keeps
// that noise away from end users. The window is just the synchronous decode call,
// during which nothing else writes to stderr.
func withQuietStderr(fn func()) {
	devNullOnce.Do(func() {
		devNull, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	})
	if devNull == nil {
		fn()
		return
	}

	stderrMu.Lock()
	saved := os.Stderr
	os.Stderr = devNull
	defer func() {
		os.Stderr = saved
		stderrMu.Unlock()
	}()
	fn()
}

// safeDecodeConfig is image.DecodeConfig hardened against the panics that the
// WASM-backed jpegli decoder raises (rather than returning an error) on malformed
// images, so the caller can simply skip an undecodable image. Decoder chatter is
// suppressed.
func safeDecodeConfig(r io.Reader) (cfg image.Config, format string, err error) {
	withQuietStderr(func() {
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("image decoder panicked: %v", rec)
			}
		}()
		cfg, format, err = image.DecodeConfig(r)
	})
	return
}

// safeOptimize runs B64Image.Optimize, returning the image unchanged if the
// optimiser errors or panics (eg. jpegli tripping over a malformed image).
func safeOptimize(img utils.B64Image) utils.B64Image {
	out := img
	withQuietStderr(func() {
		defer func() { _ = recover() }()
		if opt, changed, err := img.Optimize(); err == nil && changed {
			out = opt
		}
	})
	return out
}

func extractRecipes(e *epub.Epub, pe chan<- formats.ParseEvent) {
	docs, err := loadDocuments(e)
	if err != nil {
		pe <- formats.ParseEvent{Err: err}
		return
	}

	profile, err := induce.InduceWith(docs, bookIdent(e), modelRefiner)
	if err != nil {
		pe <- formats.ParseEvent{Err: fmt.Errorf("couldn't determine the book's recipe layout: %w", err)}
		return
	}

	report := profile.Extract(docs)
	pe <- formats.ParseEvent{N: len(report.Recipes)}
	filt := buildPhotoFilter(e)

	for _, r := range report.Recipes {
		// Flagged recipes failed the verbatim/structure gate; surface rather than ship.
		if !r.OK() {
			pe <- formats.ParseEvent{
				Err: fmt.Errorf("skipping %q: %s", r.Title, strings.Join(r.Issues, "; ")),
				I:   1,
			}
			continue
		}
		pe <- formats.ParseEvent{Recipe: toInterchange(r, e, filt), I: 1}
	}
}

// loadDocuments reads every content (X)HTML document from the ePub, skipping
// obvious front/back-matter so it doesn't pollute structure discovery.
func loadDocuments(e *epub.Epub) ([]induce.Document, error) {
	pkg, err := e.Package()
	if err != nil {
		return nil, fmt.Errorf("couldn't open ePub package file: %w", err)
	}

	var docs []induce.Document
	for _, item := range pkg.Manifest.Items {
		if !isContent(item.MediaType, item.Href) || skipped(item.Href) {
			continue
		}
		f, err := e.OpenItem(item.Href)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(f)
		if err != nil {
			continue
		}
		doc, err := induce.ParseDocument(item.Href, bytes.NewReader(data))
		if err != nil {
			continue
		}
		docs = append(docs, doc)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no readable content documents found in the ePub")
	}
	return docs, nil
}

var skipName = []string{
	"nav", "toc", "cover", "copyright", "title", "index", "contents",
	"about", "ack", "dedicat", "glossary", "frontmatter", "fm", "bm", "half",
}

func isContent(mediaType, href string) bool {
	if mediaType == "application/xhtml+xml" {
		return true
	}
	l := strings.ToLower(href)
	return strings.HasSuffix(l, ".xhtml") || strings.HasSuffix(l, ".html") || strings.HasSuffix(l, ".htm")
}

func skipped(href string) bool {
	base := strings.ToLower(path.Base(href))
	for _, p := range skipName {
		if strings.HasPrefix(base, p) {
			return true
		}
	}
	return false
}

var digits = regexp.MustCompile(`\d`)

func bookIdent(e *epub.Epub) induce.BookIdent {
	id := induce.BookIdent{}
	info, err := e.Information()
	if err != nil {
		return id
	}
	if len(info.Title) > 0 {
		id.Title = info.Title[0]
	}
	for _, ident := range info.Identifier {
		d := strings.Join(digits.FindAllString(ident.Value, -1), "")
		if len(d) == 13 || len(d) == 10 {
			id.ISBN = d
			break
		}
	}
	return id
}

func toInterchange(r induce.Recipe, e *epub.Epub, filt photoFilter) formats.Recipe {
	ir := formats.NewInterchangeRecipe()
	ir.Title = r.Title
	ir.Yield = r.Yield

	desc := r.Description
	if r.Subtitle != "" {
		if desc != "" {
			desc = r.Subtitle + "\n\n" + desc
		} else {
			desc = r.Subtitle
		}
	}
	ir.Description = desc

	for _, s := range r.Ingredients {
		ir.Ingredients = append(ir.Ingredients, formats.TitledList{Title: s.Title, List: s.Items})
	}
	for _, s := range r.Steps {
		ir.Instructions = append(ir.Instructions, formats.TitledList{Title: s.Title, List: s.Items})
	}
	ir.Tags = append(ir.Tags, r.Categories...)

	for _, p := range r.Images {
		if img, ok := loadPhoto(e, p, filt); ok {
			ir.Images = append(ir.Images, img)
			if len(ir.Images) >= maxPhotos {
				break
			}
		}
	}

	return ir
}

// loadPhoto opens a candidate image and keeps it only if it's a real photo — a
// raster at least minDim on each side and at least the book's learned byte-size
// floor — filtering out icons, the "v" glyph, chapter ornaments and diagrams.
func loadPhoto(e *epub.Epub, itemPath string, filt photoFilter) (utils.B64Image, bool) {
	f, err := e.OpenItem(itemPath)
	if err != nil {
		return nil, false
	}
	data, err := io.ReadAll(f)
	if err != nil || len(data) < filt.minBytes {
		return nil, false
	}
	cfg, format, err := safeDecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") {
		return nil, false
	}
	if cfg.Width < filt.minDim || cfg.Height < filt.minDim {
		return nil, false
	}
	return safeOptimize(utils.B64Image(data)), true
}
