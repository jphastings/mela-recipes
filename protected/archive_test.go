package protected

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/yeka/zip"
)

func bytesEntry(name string, data []byte) archiveEntry {
	return archiveEntry{name: name, write: func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	}}
}

// buildArchive writes a protected archive to an in-memory buffer.
func buildArchive(t *testing.T, entries []archiveEntry, questions, answers []string, requiredCorrect int) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := writeArchive(zw, entries, questions, answers, requiredCorrect); err != nil {
		t.Fatalf("writeArchive: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return bytes.NewReader(buf.Bytes())
}

// answerFunc answers questions from a question->answer map.
func answerFunc(m map[string]string) OwnershipTestFunc {
	return func(q string) (string, error) { return m[q], nil }
}

func noExplain(int, int) {}

var (
	roundTripQuestions = []string{"q1", "q2", "q3", "q4"}
	roundTripAnswers   = []string{"a1", "a2", "a3", "a4"}
	roundTripEntries   = []archiveEntry{
		bytesEntry("first.txt", []byte("the contents of the first file")),
		bytesEntry("second.txt", []byte("the contents of the second file")),
	}
)

func correctAnswers() OwnershipTestFunc {
	m := map[string]string{}
	for i, q := range roundTripQuestions {
		m[q] = roundTripAnswers[i]
	}
	return answerFunc(m)
}

func TestRoundTripCorrectAnswers(t *testing.T) {
	r := buildArchive(t, roundTripEntries, roundTripQuestions, roundTripAnswers, 2)

	entries, err := Read(r, r.Size(), correctAnswers(), noExplain)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(entries) != len(roundTripEntries) {
		t.Fatalf("got %d entries, want %d", len(entries), len(roundTripEntries))
	}

	want := map[string]string{
		"first.txt":  "the contents of the first file",
		"second.txt": "the contents of the second file",
	}
	for _, e := range entries {
		rc, err := e.Open()
		if err != nil {
			t.Fatalf("open %q: %v", e.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %q: %v", e.Name, err)
		}
		if string(data) != want[e.Name] {
			t.Errorf("%q = %q, want %q", e.Name, data, want[e.Name])
		}
	}
}

func TestRoundTripWrongAnswers(t *testing.T) {
	r := buildArchive(t, roundTripEntries, roundTripQuestions, roundTripAnswers, 2)

	wrong := answerFunc(map[string]string{"q1": "x", "q2": "x", "q3": "x", "q4": "x"})
	if _, err := Read(r, r.Size(), wrong, noExplain); err != ErrIncorrectAnswers {
		t.Errorf("got %v, want ErrIncorrectAnswers", err)
	}
}

func TestRoundTripTooFewAnswers(t *testing.T) {
	r := buildArchive(t, roundTripEntries, roundTripQuestions, roundTripAnswers, 2)

	none := answerFunc(map[string]string{}) // every question answered with ""
	_, err := Read(r, r.Size(), none, noExplain)
	if !errors.Is(err, ErrNotEnoughAnswers) {
		t.Errorf("got %v, want ErrNotEnoughAnswers", err)
	}
}

func TestReadNoManifest(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("loose.txt")
	w.Write([]byte("not protected"))
	zw.Close()

	r := bytes.NewReader(buf.Bytes())
	if _, err := Read(r, r.Size(), correctAnswers(), noExplain); err != ErrNoManifestFound {
		t.Errorf("got %v, want ErrNoManifestFound", err)
	}
}
