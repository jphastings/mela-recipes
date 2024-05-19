package mela

import (
	"fmt"
	"strconv"
	"strings"
)

type Book struct {
	ISBN13       string
	Pages        Pages
	RecipeNumber uint
}

func (r *Recipe) Book() *Book {
	nameString := strings.SplitN(r.ID, "#", 2)

	assignedName := strings.SplitN(nameString[0], ":", 3)
	if len(assignedName) < 3 || assignedName[0] != "urn" || assignedName[1] != "isbn" {
		return nil
	}

	isbn13, err := validateISBN(assignedName[2])
	if err != nil {
		return nil
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
					return nil
				}
			case "recipe":
				recipeNumber, err = strconv.ParseUint(keyVal[1], 10, 64)
				if err != nil {
					return nil
				}
			}
		}
	}

	if pages == nil {
		recipeNumber = 0
	}

	return &Book{
		ISBN13:       isbn13,
		Pages:        pages,
		RecipeNumber: uint(recipeNumber),
	}
}

func (r *Recipe) SetBook(isbn10or13 string, pages Pages, recipeNumber uint) error {
	isbn13, err := validateISBN(isbn10or13)
	if err != nil {
		return err
	}

	r.ID = fmt.Sprintf("urn:isbn:%s", isbn13)
	if pages != nil {
		r.ID += fmt.Sprintf("#pages=%s", pages.String())
		if recipeNumber > 0 {
			r.ID += fmt.Sprintf("&recipe=%d", recipeNumber)
		}
	}

	return nil
}

func (b *Book) String() string {
	return fmt.Sprintf("%s, p.%s", b.ISBN13, b.Pages)
}
