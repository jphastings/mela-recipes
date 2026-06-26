package protected

import (
	"fmt"
	"io"
	"os"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/yeka/zip"
)

// Option configures a Writer.
type Option func(*options)

type options struct {
	questionCount   int
	requiredCorrect int
}

// WithQuestionCount sets how many ownership questions the archive will carry.
// At least this many suitable recipes must be added. Defaults to 8.
func WithQuestionCount(n int) Option {
	return func(o *options) { o.questionCount = n }
}

// WithRequiredCorrect sets how many questions must be answered correctly to
// decrypt the archive. Defaults to 4. Must not exceed the question count.
func WithRequiredCorrect(n int) Option {
	return func(o *options) { o.requiredCorrect = n }
}

// Writer accumulates recipes and, on Close, writes them as an encrypted archive
// whose password can be derived by answering questions about their contents.
type Writer struct {
	zw      *zip.Writer
	recipes []formats.Recipe
	opts    options
}

// NewWriter returns a Writer that writes a protected archive to w.
func NewWriter(w io.Writer, opts ...Option) *Writer {
	o := options{questionCount: defaultQuestionCount, requiredCorrect: defaultRequiredCorrect}
	for _, opt := range opts {
		opt(&o)
	}
	return &Writer{zw: zip.NewWriter(w), opts: o}
}

// Add queues a recipe to be protected. The recipe (and its contents) is held in
// memory until Close, since questions can only be generated once every recipe is
// known.
func (w *Writer) Add(r formats.Recipe) error {
	w.recipes = append(w.recipes, r)
	return nil
}

// Close generates the ownership questions, derives the password, and writes the
// manifest and the encrypted recipe entries. No further recipes may be added.
func (w *Writer) Close() (err error) {
	defer func() {
		if cerr := w.zw.Close(); err == nil {
			err = cerr
		}
	}()

	irs := make([]formats.InterchangeRecipe, 0, len(w.recipes))
	for _, r := range w.recipes {
		ir, exportErr := r.Export()
		if exportErr != nil {
			return fmt.Errorf("unable to read recipe %q for question generation: %w", r.Name(), exportErr)
		}
		irs = append(irs, ir)
	}

	questions, answers, err := prepareQuestions(irs, w.opts.questionCount)
	if err != nil {
		return err
	}

	entries := make([]archiveEntry, len(w.recipes))
	for i, r := range w.recipes {
		entries[i] = archiveEntry{name: r.Filename(), write: r.Marshal}
	}

	return writeArchive(w.zw, entries, questions, answers, w.opts.requiredCorrect)
}

// archiveEntry is a single named payload to be encrypted into the archive.
type archiveEntry struct {
	name  string
	write func(io.Writer) error
}

// writeArchive writes the unencrypted manifest followed by each AES256-encrypted
// entry, deriving the password and additional points from the answers.
func writeArchive(zw *zip.Writer, entries []archiveEntry, questions, answers []string, requiredCorrect int) error {
	password, additionalPoints, err := createPassword(answers, requiredCorrect)
	if err != nil {
		return fmt.Errorf("unable to create a password for this archive: %w", err)
	}

	mw, err := zw.Create(manifestFile)
	if err != nil {
		return fmt.Errorf("unable to add manifest to archive: %w", err)
	}
	if err := writeDecryptingTXT(mw, questions, additionalPoints); err != nil {
		return fmt.Errorf("unable to write manifest into archive: %w", err)
	}

	for _, e := range entries {
		ew, err := zw.Encrypt(e.name, password, zip.AES256Encryption)
		if err != nil {
			return fmt.Errorf("unable to create encrypted entry %q: %w", e.name, err)
		}
		if err := e.write(ew); err != nil {
			return fmt.Errorf("unable to write entry %q into archive: %w", e.name, err)
		}
	}

	return nil
}

// Create writes the given recipes to a protected archive at path.
func Create(path string, recipes []formats.Recipe, opts ...Option) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	w := NewWriter(f, opts...)
	for _, r := range recipes {
		if err = w.Add(r); err != nil {
			return err
		}
	}
	return w.Close()
}
