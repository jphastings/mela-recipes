package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type Book struct {
	ISBN13 string
	Title  string
}

type BookRef struct {
	Book
	Pages        Pages
	RecipeNumber uint
}

func NewBookRefFromURN(urn string) (BookRef, error) {
	nameString := strings.SplitN(urn, "#", 2)

	assignedName := strings.SplitN(nameString[0], ":", 3)
	if len(assignedName) < 3 || assignedName[0] != "urn" || assignedName[1] != "isbn" {
		return BookRef{}, fmt.Errorf("provided string is not an ISBN URN")
	}

	isbn13, err := StandardizeISBN(assignedName[2])
	if err != nil {
		return BookRef{}, fmt.Errorf("the ISBN within this URN is invalid: %w", err)
	}

	var pages Pages
	var recipeNumber uint64

	if len(nameString) == 2 {
		// Custom Query param pasing, as we don't want to url decode the whole string
		fragments := strings.Split(nameString[1], "&")
		for _, fragment := range fragments {
			keyVal := strings.SplitN(fragment, "=", 2)
			if len(keyVal) != 2 {
				continue
			}

			switch keyVal[0] {
			case "pages":
				pages, err = ParsePages(keyVal[1])
				if err != nil {
					return BookRef{}, fmt.Errorf("the pages reference within th ISBN URN couldn't be parsed: %w", err)
				}
			case "recipe":
				recipeNumber, err = strconv.ParseUint(keyVal[1], 10, 64)
				if err != nil {
					return BookRef{}, fmt.Errorf("the recipe number reference within th ISBN URN couldn't be parsed: %w", err)
				}
			}
		}
	}

	if pages == nil {
		recipeNumber = 0
	}

	return BookRef{
		Book: Book{
			ISBN13: isbn13,
		},
		Pages:        pages,
		RecipeNumber: uint(recipeNumber),
	}, nil
}

func (br BookRef) URN() string {
	urn := fmt.Sprintf("urn:isbn:%s", br.Book.ISBN13)
	if br.Pages == nil {
		return urn
	}

	urn += fmt.Sprintf("#pages=%s", br.Pages.String())
	if br.RecipeNumber == 0 {
		return urn
	}

	urn += fmt.Sprintf("&recipe=%d", br.RecipeNumber)
	return urn
}

func (b *Book) String() string {
	if b.Title == "" {
		return fmt.Sprintf("%s", b.ISBN13)
	} else {
		return fmt.Sprintf("%s (%s)", b.Title, b.ISBN13)
	}
}

func (br *BookRef) String() string {
	return fmt.Sprintf("%s, p.%s", br.Book, br.Pages)
}
