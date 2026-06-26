package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
)

// withOwnershipPrompts attaches interactive terminal callbacks for answering the
// proof-of-ownership questions of a .protectedrecipes file. Questions are written
// to stderr (so stdout stays clean for piping) and answers are read, echoed, from
// stdin — they're book content, not secrets.
func withOwnershipPrompts(o formats.ParseOptions) formats.ParseOptions {
	r := bufio.NewReader(os.Stdin)

	o.ExplainOwnership = func(needed, mayFail int) {
		fmt.Fprintf(os.Stderr, "🔒 These recipes are protected. Answer %d question(s) about the book to unlock them (you may skip or get up to %d wrong). Press enter to skip a question.\n\n", needed, mayFail)
	}

	o.AskOwnership = func(question string) (string, error) {
		fmt.Fprintf(os.Stderr, "%s\n> ", question)
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}

	return o
}
