package protected

import (
	"bufio"
	"fmt"
	"io"
)

const (
	// manifestFile is the name of the unencrypted entry, stored first in a
	// protected archive, that holds the ownership questions and the additional
	// points needed to derive the decryption password.
	manifestFile = "_decrypting.txt"

	// decryptingExplainer separates the human-answerable questions from the
	// machine-only additional points within the manifest.
	decryptingExplainer = "# Above are questions that allow the derivation of the password for the other files in this archive, below is additional machine information needed for the same. Please see https://github.com/jphastings/recipes/tree/main/protected for specifics."
)

// writeDecryptingTXT writes the manifest (questions, then the explainer comment,
// then the additional points) to w.
func writeDecryptingTXT(w io.Writer, questions []string, additionalPoints []string) error {
	for _, q := range questions {
		if _, err := fmt.Fprintln(w, q); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w, decryptingExplainer); err != nil {
		return err
	}

	for _, a := range additionalPoints {
		if _, err := fmt.Fprintln(w, a); err != nil {
			return err
		}
	}

	return nil
}

// readDecryptingTXT parses the manifest, splitting the questions from the
// additional points at the explainer comment line.
func readDecryptingTXT(r io.Reader) (questions, additionalPoints []string, err error) {
	fileScanner := bufio.NewScanner(r)
	fileScanner.Split(bufio.ScanLines)

	readingQuestions := true
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if len(line) > 0 && line[0] == '#' {
			readingQuestions = false
			continue
		}

		if readingQuestions {
			questions = append(questions, line)
		} else {
			additionalPoints = append(additionalPoints, line)
		}
	}

	if err := fileScanner.Err(); err != nil {
		return nil, nil, err
	}

	return questions, additionalPoints, nil
}
