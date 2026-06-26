package protected

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"

	"github.com/yeka/zip"
)

// DecryptedEntry is a single decrypted file from a protected archive. Open lazily
// decrypts and returns the entry's contents; the underlying reader passed to Read
// (or the Closer returned by Open) must remain open until reading is finished.
type DecryptedEntry struct {
	Name string
	Open func() (io.ReadCloser, error)
}

// Read opens a protected archive from r, reads its question manifest, and (via the
// callbacks) collects answers to derive the decryption password. It returns the
// decrypted entries; the caller is responsible for parsing them. The password is
// determined once, lazily, against the first encrypted entry.
func Read(r io.ReaderAt, size int64, onTest OwnershipTestFunc, onExplain OwnershipExplainFunc) ([]DecryptedEntry, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("unable to open zip: %w", err)
	}

	questions, additionalPoints, err := readManifest(zr)
	if err != nil {
		return nil, err
	}

	var password string
	var passwordErr error
	var once sync.Once
	ensurePassword := func(zf *zip.File) error {
		once.Do(func() {
			password, passwordErr = determinePassword(questions, additionalPoints, onTest, onExplain, testPasswordOn(zf))
		})
		return passwordErr
	}

	var entries []DecryptedEntry
	for _, zf := range zr.File {
		if zf.Name == manifestFile || !zf.IsEncrypted() {
			continue
		}

		if err := ensurePassword(zf); err != nil {
			return nil, err
		}
		zf.SetPassword(password)

		zf := zf
		entries = append(entries, DecryptedEntry{
			Name: zf.Name,
			Open: func() (io.ReadCloser, error) { return zf.Open() },
		})
	}

	return entries, nil
}

// Open is a convenience wrapper around Read for a file on disk. The returned
// Closer must be kept open until the entries have been read, then closed.
func Open(path string, onTest OwnershipTestFunc, onExplain OwnershipExplainFunc) ([]DecryptedEntry, io.Closer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	entries, err := Read(f, fi.Size(), onTest, onExplain)
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	return entries, f, nil
}

// readManifest finds and parses the unencrypted question manifest.
func readManifest(zr *zip.Reader) (questions, additionalPoints []string, err error) {
	for _, zf := range zr.File {
		if zf.Name != manifestFile {
			continue
		}

		rr, err := zf.Open()
		if err != nil {
			return nil, nil, fmt.Errorf("unable to open manifest: %w", err)
		}
		defer rr.Close()

		questions, additionalPoints, err = readDecryptingTXT(rr)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse manifest: %w", err)
		}
		return questions, additionalPoints, nil
	}

	return nil, nil, ErrNoManifestFound
}

// passwordTester reports whether a candidate password is correct.
type passwordTester func(string) (bool, error)

// determinePassword asks questions (in a random order) until enough are answered
// to derive a candidate password, then verifies it. It does not retry alternative
// answer combinations: if the derived password is wrong, at least one answer was.
func determinePassword(questions, additionalPoints []string, onTest OwnershipTestFunc, onExplain OwnershipExplainFunc, isPasswordOK passwordTester) (string, error) {
	needAnswerCount := len(questions) - len(additionalPoints)
	qIDs := rand.Perm(len(questions))

	answers := make(map[int]string)
	var errs error
	for i, x := range qIDs {
		if len(answers)+len(qIDs)-i < needAnswerCount {
			break
		}

		if i == 0 {
			onExplain(needAnswerCount, len(additionalPoints))
		}

		ans, err := onTest(questions[x])
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("unable to collect answer to question %d: %w", x, err))
			continue
		}
		if ans == "" {
			continue
		}

		answers[x] = ans
		if len(answers) >= needAnswerCount {
			break
		}
	}
	if len(answers) < needAnswerCount {
		return "", errors.Join(ErrNotEnoughAnswers, errs)
	}

	password, err := derivePassword(answers, len(questions), additionalPoints)
	if err != nil {
		return "", fmt.Errorf("unable to derive password: %w", err)
	}

	ok, err := isPasswordOK(password)
	if err != nil {
		return "", fmt.Errorf("the derived password could not be tested: %w", err)
	}
	if !ok {
		return "", ErrIncorrectAnswers
	}

	return password, nil
}

// testPasswordOn returns a tester that checks a candidate password against zf.
func testPasswordOn(zf *zip.File) passwordTester {
	return func(possiblePassword string) (bool, error) {
		zf.SetPassword(possiblePassword)

		rc, err := zf.Open()
		switch {
		case err == nil:
			rc.Close()
			return true, nil
		case errors.Is(err, zip.ErrPassword):
			return false, nil
		default:
			return false, err
		}
	}
}
