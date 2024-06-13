package mela

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"strings"

	interpolation "github.com/SadPencil/go-lagrange-interpolation"
	"github.com/SadPencil/go-lagrange-interpolation/field"
	"github.com/yeka/zip"
	"golang.org/x/text/unicode/norm"
)

type ProtectedRecipes struct {
	zip  *zip.Writer
	isbn string

	unprotectedRecipes []*Recipe
}

// OwnershipTestFunc is called if the recipes file being parsed requires proof of ownership to provide the details to decrypt.
// A single question will be provided as the argument, and the answer should be returned. (It will be downcased, whitespace trimmed, and NFKC normalized)
type OwnershipTestFunc func(string) (string, error)

var (
	encMethod           = zip.AES256Encryption
	questionsFile       = "_decrypting.txt"
	decryptingExplainer = "# Above are questions that allow the derivation of the password for the other files in this archive, below is additional machine information needed for the same. Please see https://github.com/jphastings/mela-recipes#proof-of-ownership-extension for specifics."
	m127                *big.Int
)

func init() {
	var ok bool
	m127, ok = big.NewInt(0).SetString("170141183460469231731687303715884105727", 10)
	if !ok {
		panic("The Mersenne-127 prime number has been corrupted")
	}
}

func newProtectedRecipes(w io.Writer) *ProtectedRecipes {
	return &ProtectedRecipes{
		zip: zip.NewWriter(w),
	}
}

// ParseProtectedRecipes parses a known .protectedrecipes collection file (or .protectedrecipes file) into a stream of Recipe-compatible structs, calling the onRecipe func for each, as it is parsed.
func ParseProtectedRecipes(r io.ReaderAt, size int64, onRecipe RecipeFunc, onOwnershipTest OwnershipTestFunc) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return fmt.Errorf("unable to open zip; %w", err)
	}

	var password string
	for _, zf := range zr.File {
		if zf.Name != questionsFile {
			continue
		}

		rr, err := zf.Open()
		if err != nil {
			return fmt.Errorf("unable to open questions file from archive: %w", err)
		}
		defer rr.Close()

		questions, additionalPoints, err := readDecryptingTXT(rr)
		if err != nil {
			return fmt.Errorf("unable to parse questions file: %w", err)
		}

		qIDs, err := pickNofM(len(questions)-len(additionalPoints), len(questions))
		if err != nil {
			return err
		}

		// TODO: Break this out so I can query for more answers if the password is incorrect
		answers := make(map[int]string)
		for _, x := range qIDs {
			ans, err := onOwnershipTest(questions[x])
			if err != nil {
				return fmt.Errorf("unable to collect answer to question: %w", err)
			}

			answers[x] = ans
		}

		password, err = derivePassword(answers, len(questions), additionalPoints)
		if err != nil {
			return fmt.Errorf("unable to derive password: %w", err)
		}
	}

	if password == "" {
		return ErrNoQuestionsFileFound
	}

	for _, zf := range zr.File {
		if !strings.HasSuffix(zf.Name, ".melarecipe") {
			continue
		}
		if zf.IsEncrypted() {
			zf.SetPassword(password)
		}

		rr, err := zf.Open()
		if err != nil {
			if errors.Is(err, zip.ErrPassword) {
				return ErrIncorrectAnswers
			}

			onRecipe(nil, err)
		}
		defer rr.Close()

		if recipe, err := ParseRecipe(rr); err != nil {
			onRecipe(nil, err)
		} else {
			recipe.Filename = withoutExt(zf.Name)
			onRecipe(recipe, nil)
		}
	}

	return nil
}

func (pr *ProtectedRecipes) Close() error {
	defer pr.zip.Close()

	// TODO: Generate questions & answers
	questions := []string{
		"Q1?", "Q2?", "Q3?", "Q4?", "Q5?", "Q6?", "Q7?",
	}
	answers := []string{
		"flambé", "A2", "A3", "A4", "A5", "A6", "A7",
	}

	password, additionalPoints, err := createPassword(answers, 3)
	if err != nil {
		return fmt.Errorf("unable to create a password suitable for this recipe: %w", err)
	}

	w, err := pr.zip.Create(questionsFile)
	if err != nil {
		return fmt.Errorf("unable to add decryption explainer to zip file: %w", err)
	}
	if err := writeDecryptingTXT(w, questions, additionalPoints); err != nil {
		return fmt.Errorf("unable to write decryption explainer into zip file: %w", err)
	}

	for _, r := range pr.unprotectedRecipes {
		w, err := pr.zip.Encrypt(r.Filename+".melarecipe", password, encMethod)
		if err != nil {
			return err
		}

		if err := json.NewEncoder(w).Encode(r); err != nil {
			return fmt.Errorf("unable to encode recipe JSON into encrypted zip: %w", err)
		}
	}

	if err := pr.zip.Flush(); err != nil {
		return err
	}
	return nil
}

func (pr *ProtectedRecipes) Add(r *Recipe) error {
	if r.Book().ISBN13 == "" {
		return fmt.Errorf("only recipes with an ISBN can be added to a protected recipes bundle")
	}

	if len(pr.unprotectedRecipes) == 0 {
		pr.isbn = r.Book().ISBN13
	} else if pr.isbn != r.Book().ISBN13 {
		return fmt.Errorf("all recipes added to a protected recipes bundle must be from the same book (and have the same ISBN)")
	}

	pr.unprotectedRecipes = append(pr.unprotectedRecipes, r)
	return nil
}

// normalize normalizes text so that simple reading/inputting dissimilarities are corrected for
func normalize(text string) string {
	return norm.NFKC.String(strings.ToLower(strings.TrimSpace(text)))
}

var (
	ErrNotEnoughAnswers         = errors.New("at least 2 answers must be provided")
	ErrNotEnoughRequiredCorrect = errors.New("at one answer must be required correct")
	ErrTooManyRequiredCorrect   = errors.New("more answers are required to be correct than are provided")
	ErrInvalidAnswers           = errors.New("the answers provided could not be used to produce a password")
	ErrNoQuestionsFileFound     = errors.New("no questions file found in the archive, is this a protected recipes file?")
	ErrIncorrectAnswers         = errors.New("not enough of the provided answers were correct to decrypt this recipes file")
)

// createPassword calculates the password, as well as the additional points needed for deriving it from the provided answers. Non-normalized answers can be provided.
func createPassword(answers []string, requiredCorrect int) (string, []string, error) {
	ansCount := len(answers)
	if ansCount < 2 {
		return "", nil, ErrNotEnoughAnswers
	}
	if requiredCorrect < 1 {
		return "", nil, ErrNotEnoughAnswers
	}
	if requiredCorrect > ansCount {
		return "", nil, ErrTooManyRequiredCorrect
	}

	var password string
	var additionalPoints []string

	var points []*interpolation.XYPoint
	for x, ans := range answers {
		h := sha256.Sum256([]byte(normalize(ans)))
		points = append(points, bytesToPoint(x, h[:]))
	}

	poly, err := interpolation.LagrangeInterpolation(points)
	if err != nil {
		return "", nil, ErrInvalidAnswers
	}

	for x := ansCount; x <= ansCount*2-requiredCorrect; x++ {
		y := poly.EvaluateAt(intField(x))
		b64 := base64.RawStdEncoding.EncodeToString(y.Value.Bytes())
		if x == ansCount {
			password = b64
		} else {
			additionalPoints = append(additionalPoints, b64)
		}
	}

	return password, additionalPoints, nil
}

// derivePassword reverse engineers the password from provided answers and the additional point data needed
func derivePassword(answers map[int]string, passwordX int, additionalPoints []string) (string, error) {
	if len(answers)+len(additionalPoints) != passwordX {
		return "", fmt.Errorf("incorrect number of answers provided: got %d, need %d", len(answers), passwordX)
	}

	var points []*interpolation.XYPoint
	for x, ans := range answers {
		h := sha256.Sum256([]byte(normalize(ans)))
		points = append(points, bytesToPoint(x, h[:]))
	}
	for i, a := range additionalPoints {
		b, err := base64.RawStdEncoding.DecodeString(a)
		if err != nil {
			return "", fmt.Errorf("invalid additional point could not be base64 decoded: %s", a)
		}

		points = append(points, bytesToPoint(passwordX+i+1, b))
	}

	poly, err := interpolation.LagrangeInterpolation(points)
	if err != nil {
		return "", ErrInvalidAnswers
	}

	y := poly.EvaluateAt(intField(passwordX))
	return base64.RawStdEncoding.EncodeToString(y.Value.Bytes()), nil
}

// bytesToPoint is a helper function that encodes a (big endian) byte slice representing an integer as the Y value of an XYPoint in the M-127 field
func bytesToPoint(x int, b []byte) *interpolation.XYPoint {
	y := big.NewInt(0).SetBytes(b)
	return &interpolation.XYPoint{
		X: intField(x),
		Y: &field.Field{Modulus: m127, Value: y},
	}
}

// writeDecryptingTXT writes the contents of the decrypting guide file to the given writer
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

// readDecryptingTXT reads the contents of the decrypting guide file and extracts the useful data
func readDecryptingTXT(r io.Reader) (questions, additionalPoints []string, err error) {
	fileScanner := bufio.NewScanner(r)
	fileScanner.Split(bufio.ScanLines)

	readingQuestions := true
	for fileScanner.Scan() {
		if line := fileScanner.Text(); line[0] == '#' {
			readingQuestions = false
		} else {
			if readingQuestions {
				questions = append(questions, line)
			} else {
				additionalPoints = append(additionalPoints, line)
			}
		}
	}

	if err := fileScanner.Err(); err != nil {
		return nil, nil, err
	}

	return questions, additionalPoints, nil
}

// intField is a shorthand for producing a field with the correct modulus from an int
func intField(n int) *field.Field {
	return &field.Field{Modulus: m127, Value: big.NewInt(int64(n))}
}

// pickNofM randomly selects n elements from 0 up to (but not including) m with no repeating selections.
func pickNofM(n, m int) ([]int, error) {
	if n > m {
		return nil, fmt.Errorf("cannot pick; there are not %d items from 0 up to (but excluding) %d", n, m)
	}

	list := rand.Perm(m)
	return list[:n], nil
}
