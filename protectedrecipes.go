package mela

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/yeka/zip"
)

type ProtectedRecipes struct {
	zip  *zip.Writer
	isbn string

	unprotectedRecipes []*Recipe
}

// OwnershipTestFunc is called if the recipes file being parsed requires proof of ownership to provide the details to decrypt.
// A single question will be provided as the argument, and the answer should be returned. (It will be downcased, whitespace trimmed, and NFKC normalized)
type OwnershipTestFunc func(string) string

func newProtectedRecipes(w io.Writer) *ProtectedRecipes {
	return &ProtectedRecipes{
		zip: zip.NewWriter(w),
	}
}

// ParseProtectedRecipes parses a known .protectedrecipes collection file (or .protectedrecipes file) into a stream of Recipe-compatible structs, calling the onRecipe func for each, as it is parsed.
func ParseProtectedRecipes(r io.ReaderAt, size int64, onRecipe RecipeFunc, onOwnershipTest OwnershipTestFunc) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}

	for _, zf := range zr.File {
		rr, err := zf.Open()
		if err != nil {
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

var (
	encMethod           = zip.AES256Encryption
	questionsFile       = "_decrypting.txt"
	decryptingExplainer = "# These are questions that allow the derivation of the password for the other files in this archive. Please see https://github.com/jphastings/mela-recipes#proof-of-ownership-extension for information."
)

func (pr *ProtectedRecipes) Close() error {
	w, err := pr.zip.Create(questionsFile)
	if err != nil {
		return fmt.Errorf("unable to add explainer to zip file: %w", err)
	}

	// TODO: Generate questions & password
	questions := [7]string{
		"Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7",
	}
	password := "5+zQnJ2h6uSj/o4qLAuhzg"

	for _, q := range questions {
		fmt.Fprintln(w, q)
	}

	fmt.Fprintln(w, decryptingExplainer)

	for _, r := range pr.unprotectedRecipes {
		w, err := pr.zip.Encrypt(r.Filename+".melarecipe", password, encMethod)
		if err != nil {
			return err
		}

		if err := json.NewEncoder(w).Encode(r); err != nil {
			return fmt.Errorf("unable to encode recipe JSON into encrypted zip: %w", err)
		}
	}

	return pr.zip.Close()
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
