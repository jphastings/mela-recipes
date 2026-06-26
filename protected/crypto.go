// Package protected reads and writes "protected" archives: zip files whose
// entries are AES256 encrypted with a password that can only be reconstructed by
// answering questions about their contents. It is used to bundle recipes derived
// from a physical book such that only someone able to consult that book (and so,
// presumably, an owner of it) can decrypt them.
//
// The password is shared across the questions using Lagrange interpolation over
// the field of the 127th Mersenne prime, so that only a subset of the questions
// need be answered correctly. See the accompanying manifest format in
// _decrypting.txt for the on-disk specifics.
package protected

import (
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"strings"

	interpolation "github.com/SadPencil/go-lagrange-interpolation"
	"github.com/SadPencil/go-lagrange-interpolation/field"
	"golang.org/x/text/unicode/norm"
)

// m127 is the 127th Mersenne prime, used as the modulus of the field in which
// the password is shared and reconstructed.
var m127 *big.Int

func init() {
	var ok bool
	m127, ok = big.NewInt(0).SetString("170141183460469231731687303715884105727", 10)
	if !ok {
		panic("the Mersenne-127 prime number has been corrupted")
	}
}

// normalize normalizes text so that simple reading/inputting dissimilarities are
// corrected for, allowing answers typed by a human to match the originals.
func normalize(text string) string {
	return norm.NFKC.String(strings.ToLower(strings.TrimSpace(text)))
}

// createPassword calculates the password, along with the additional points needed
// to derive it from requiredCorrect of the provided answers. Non-normalized
// answers can be provided.
func createPassword(answers []string, requiredCorrect int) (string, []string, error) {
	ansCount := len(answers)
	if ansCount < 2 {
		return "", nil, ErrNotEnoughAnswers
	}
	if requiredCorrect < 1 {
		return "", nil, ErrNotEnoughRequiredCorrect
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

// derivePassword reverse engineers the password from provided answers and the
// additional point data needed.
func derivePassword(answers map[int]string, passwordX int, additionalPoints []string) (string, error) {
	if len(answers)+len(additionalPoints) < passwordX {
		return "", ErrNotEnoughAnswers
	}

	var points []*interpolation.XYPoint
	for x, ans := range answers {
		h := sha256.Sum256([]byte(normalize(ans)))
		points = append(points, bytesToPoint(x, h[:]))
	}
	for i, a := range additionalPoints {
		b, err := base64.RawStdEncoding.DecodeString(a)
		if err != nil {
			return "", ErrInvalidAnswers
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

// bytesToPoint encodes a (big endian) byte slice representing an integer as the Y
// value of an XYPoint in the M-127 field.
func bytesToPoint(x int, b []byte) *interpolation.XYPoint {
	y := big.NewInt(0).SetBytes(b)
	return &interpolation.XYPoint{
		X: intField(x),
		Y: &field.Field{Modulus: m127, Value: y},
	}
}

// intField is shorthand for producing a field with the correct modulus from an int.
func intField(n int) *field.Field {
	return &field.Field{Modulus: m127, Value: big.NewInt(int64(n))}
}
