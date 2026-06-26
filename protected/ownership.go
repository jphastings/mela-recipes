package protected

import "errors"

// OwnershipTestFunc is called when a protected archive requires proof of ownership
// to decrypt. A single question is provided, and the answer should be returned. It
// will be whitespace-trimmed, lowercased, and NFKC normalized before use. Returning
// an empty answer (with no error) is treated as "skip this question".
type OwnershipTestFunc func(question string) (answer string, err error)

// OwnershipExplainFunc is called once, before the first question is asked, so the
// caller can set context for what's expected. questionCount is the number of
// questions that must be answered correctly; failCount is how many may be skipped
// or answered incorrectly while still allowing decryption.
type OwnershipExplainFunc func(questionCount, failCount int)

var (
	ErrNotEnoughAnswers         = errors.New("not enough answers have been provided")
	ErrNotEnoughRequiredCorrect = errors.New("at least one answer must be required correct")
	ErrTooManyRequiredCorrect   = errors.New("more answers are required to be correct than are provided")
	ErrInvalidAnswers           = errors.New("the answers provided could not be used to produce a password")
	ErrNoManifestFound          = errors.New("no _decrypting.txt manifest found in the archive; is this a protected file?")
	ErrIncorrectAnswers         = errors.New("not enough of the provided answers were correct to decrypt this archive")
)
