package utils

import (
	"fmt"
	"regexp"
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

var bookFromNotesExtractor = regexp.MustCompile(`(?i)(\s*)((?:isbn:? ?|_)([0-9X-]+)\r?\n?((?:, p\.|pages?:? ?)([^_\s,]+)\r?\n?((?:recipe:? ?|, )?(\d+)(?:[a-z]{2})?\r?\n?)?)?)_?(\s*)`)

// ExtractBookFromNotes finds a book reference embedded in a recipe's notes (eg.
// "ISBN: 978-3-16-148410-0\npages: 52") and rewrites it into the canonical
// "_<isbn>, p.<pages>, <n>th_" form. It returns the rewritten notes, the parsed
// reference, and whether one was found; found is false (with no error) when the
// notes contain no recognisable reference, in which case notes is returned
// unchanged.
func ExtractBookFromNotes(notes string) (string, BookRef, bool, error) {
	matches := bookFromNotesExtractor.FindStringSubmatch(notes)
	if matches == nil {
		return notes, BookRef{}, false, nil
	}

	var newNotes string
	around := strings.SplitN(notes, matches[0], 2)
	if around[0] == "" {
		newNotes = around[1]
		if around[1] != "" {
			newNotes += "\n\n"
		}
	} else if around[1] == "" {
		newNotes = around[0] + "\n\n"
	} else {
		newNotes = around[0] + matches[1] + around[1] + "\n\n"
	}

	isbn13, err := StandardizeISBN(matches[3])
	if err != nil {
		return notes, BookRef{}, false, err
	}

	newNotes += fmt.Sprintf("_%s", isbn13)

	var pages Pages
	var recipeNumber uint64

	if matches[5] != "" {
		pages, err = ParsePages(matches[5])
		if err != nil {
			return notes, BookRef{}, false, err
		}

		newNotes += fmt.Sprintf(", p.%s", pages.String())
	}

	if matches[7] != "" && pages != nil {
		recipeNumber, err = strconv.ParseUint(matches[7], 10, 64)
		if err != nil {
			return notes, BookRef{}, false, err
		}

		newNotes += fmt.Sprintf(", %s", Ordinal(recipeNumber, false))
	}

	newNotes += "_"

	book := BookRef{
		Book:         Book{ISBN13: isbn13},
		Pages:        pages,
		RecipeNumber: uint(recipeNumber),
	}

	return newNotes, book, true, nil
}
