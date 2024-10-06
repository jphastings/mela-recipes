package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/jphastings/recipes/mela"
)

var (
	version = "0.0.0"
	commit  = "dev"
	date    = time.Now().Format(time.DateOnly)
)

const explainer = `Standardizes Mela recipe files (https://mela.recipes) in a number of ways:
- Reduces image sizes to a maximum of 512x512 pixels.
- Encodes images in the highly efficient JPEGli codec.
- Extracts an ISBN, page number & recipe number from the notes field, if they
  exist (eg. ISBN: 978-3-16-148410-0; or 9783161484100, p.52-55, 2nd) and
	stores them in the ID too.
- (If an ISBN was extracted) Pulls the book title from the Open Library, if possible
`

func main() {
	if len(os.Args) < 3 {
		execName := filepath.Base(os.Args[0])
		fmt.Printf(
			"Mela Standardize v%s-%s (%s)\n"+
				"Usage: %s <.melarecipe(s)> [...<.melarecipe(s)>,<.protectedrecipes>] <output directory>\n\n"+
				explainer,
			version, commit, date, execName)
		os.Exit(1)
	}

	inputFiles := os.Args[1 : len(os.Args)-1]
	outputDir := os.Args[len(os.Args)-1]

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Output directory '%s' does not exist\n", outputDir)
		os.Exit(1)
	}

	for _, file := range inputFiles {
		fmt.Printf("📚 Extracting & standardizing %s\n", path.Base(file))

		recipes, err := mela.Open(file, mela.OpenProtectedRecipes(questionHelp, getAnswer))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening '%s': %v\n", file, err)
			os.Exit(1)
		}

		for _, r := range recipes {
			if err := r.Standardize(true); err != nil {
				fmt.Fprintf(os.Stderr, "Error standardizing '%s' from '%s': %v\n", r.Title, file, err)
				os.Exit(1)
			}

			for _, s := range r.ListStandardizations() {
				fmt.Printf("→ %s\n", s)
			}

			dest, err := r.Save(outputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error saving '%s' from '%s': %v\n", r.Title, file, err)
				os.Exit(1)
			}

			fmt.Printf("Saved '%s' to '%s'\n", r.Title, dest)
		}
	}
}

func questionHelp(bookName string, questionCount, failCount int) {
	fmt.Printf("❗️ This recipe bundle contains recipes from %s. To extract them you will need a copy of this book.\n", bookName)
	fmt.Printf("✅ You must answer %d of the following questions correctly to extract recipes from it.\n", questionCount)
	if failCount > 0 {
		var upto string
		switch failCount {
		case 1:
			upto = "once"
		case 2:
			upto = "up to twice"
		default:
			upto = fmt.Sprintf("up to %d times", failCount)
		}
		fmt.Printf("🤔 You can skip a question %s by pressing return.\n", upto)
	}
	fmt.Println()
}

// getAnswer uses the CLI to ask a question about a protected recipe book, and retrieve an answer from stdin
func getAnswer(q string) (string, error) {
	fmt.Println("Q:", q)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("couldn't get input")
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error while reading input: %w", err)
	}

	fmt.Println()

	return scanner.Text(), nil
}
