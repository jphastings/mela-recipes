package protected

import (
	"bytes"
	"testing"
)

// answers used across the crypto tests; their content is irrelevant, only that
// they are distinct.
var testAnswers = []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel"}

// subset builds the answers map derivePassword expects from a set of indices.
func subset(answers []string, indices ...int) map[int]string {
	m := make(map[int]string, len(indices))
	for _, i := range indices {
		m[i] = answers[i]
	}
	return m
}

func TestPasswordRecoverableFromAnyRequiredSubset(t *testing.T) {
	const requiredCorrect = 4

	password, additionalPoints, err := createPassword(testAnswers, requiredCorrect)
	if err != nil {
		t.Fatalf("createPassword: %v", err)
	}
	if got, want := len(additionalPoints), len(testAnswers)-requiredCorrect; got != want {
		t.Fatalf("got %d additional points, want %d", got, want)
	}

	// Any requiredCorrect correct answers, plus the additional points, must
	// reconstruct the same password.
	for _, indices := range [][]int{
		{0, 1, 2, 3},
		{4, 5, 6, 7},
		{0, 2, 4, 6},
		{1, 3, 5, 7},
		{7, 0, 4, 2},
	} {
		got, err := derivePassword(subset(testAnswers, indices...), len(testAnswers), additionalPoints)
		if err != nil {
			t.Fatalf("derivePassword(%v): %v", indices, err)
		}
		if got != password {
			t.Errorf("derivePassword(%v) = %q, want %q", indices, got, password)
		}
	}
}

func TestWrongAnswerDerivesWrongPassword(t *testing.T) {
	password, additionalPoints, err := createPassword(testAnswers, 4)
	if err != nil {
		t.Fatalf("createPassword: %v", err)
	}

	answers := subset(testAnswers, 0, 1, 2, 3)
	answers[3] = "not the real answer"

	got, err := derivePassword(answers, len(testAnswers), additionalPoints)
	if err != nil {
		t.Fatalf("derivePassword: %v", err)
	}
	if got == password {
		t.Error("a wrong answer derived the correct password")
	}
}

func TestCreatePasswordValidation(t *testing.T) {
	tests := []struct {
		name            string
		answers         []string
		requiredCorrect int
		wantErr         error
	}{
		{"too few answers", []string{"only one"}, 1, ErrNotEnoughAnswers},
		{"no required correct", testAnswers, 0, ErrNotEnoughRequiredCorrect},
		{"more required than answers", testAnswers, len(testAnswers) + 1, ErrTooManyRequiredCorrect},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := createPassword(tc.answers, tc.requiredCorrect); err != tc.wantErr {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestManifestRoundTrip(t *testing.T) {
	questions := []string{"What is the title?", "What is the second word?"}
	points := []string{"S7yj6S74aoH+OHe13fSZyA", "a5D0f5T4cNy6WsVHs9YXGw"}

	var buf bytes.Buffer
	if err := writeDecryptingTXT(&buf, questions, points); err != nil {
		t.Fatalf("writeDecryptingTXT: %v", err)
	}

	gotQ, gotP, err := readDecryptingTXT(&buf)
	if err != nil {
		t.Fatalf("readDecryptingTXT: %v", err)
	}

	if len(gotQ) != len(questions) {
		t.Fatalf("got %d questions, want %d", len(gotQ), len(questions))
	}
	for i := range questions {
		if gotQ[i] != questions[i] {
			t.Errorf("question %d = %q, want %q", i, gotQ[i], questions[i])
		}
	}
	for i := range points {
		if gotP[i] != points[i] {
			t.Errorf("point %d = %q, want %q", i, gotP[i], points[i])
		}
	}
}
