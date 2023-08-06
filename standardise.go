package mela

import (
	"regexp"
	"strconv"
	"strings"
)

func (r *RawRecipe) Standardize() Recipe {
	book, notes := bookFromNotes(r.RawNotes)
	if book != nil {
		r.SetBook(book.ISBN13, book.Pages, book.RecipeNumber)
		r.RawNotes = notes
	}
	return r
}

var extractor = regexp.MustCompile(`(?i)(\s*)(isbn:? ?([0-9X-]+)\r?\n?(pages?:? ?([0-9a-z-,%]+)\r?\n?(recipe:? ?(\d+)\r?\n?)?)?)(\s*)`)

func bookFromNotes(notes string) (*Book, string) {
	matches := extractor.FindStringSubmatch(notes)
	if matches == nil {
		return nil, notes
	}

	var newNotes string
	around := strings.SplitN(notes, matches[0], 2)
	if around[0] == "" {
		newNotes = around[1]
	} else if around[1] == "" {
		newNotes = around[0]
	} else {
		newNotes = around[0] + matches[1] + around[1]
	}

	isbn13, err := validateISBN(matches[3])
	if err != nil {
		return nil, notes
	}

	var pages Pages
	var recipeNumber uint64

	if matches[5] != "" {
		pages, err = ParsePages(matches[5])
		if err != nil {
			return nil, notes
		}
	}

	if matches[7] != "" {
		recipeNumber, err = strconv.ParseUint(matches[7], 10, 64)
		if err != nil {
			return nil, notes
		}
	}

	book := &Book{
		ISBN13:       isbn13,
		Pages:        pages,
		RecipeNumber: uint(recipeNumber),
	}

	return book, newNotes
}
