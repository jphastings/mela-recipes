package mela

import (
	"fmt"
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

var extractor = regexp.MustCompile(`(?i)(\s*)((?:isbn:? ?|_)([0-9X-]+)\r?\n?((?:, p.|pages?:? ?)([^_\s,]+)\r?\n?((?:recipe:? ?|, )?(\d+)(?:[a-z]{2})?\r?\n?)?)?)_?(\s*)`)

func bookFromNotes(notes string) (*Book, string) {
	matches := extractor.FindStringSubmatch(notes)
	if matches == nil {
		return nil, notes
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

	isbn13, err := validateISBN(matches[3])
	if err != nil {
		return nil, notes
	}

	newNotes += fmt.Sprintf("_%s", isbn13)

	var pages Pages
	var recipeNumber uint64

	if matches[5] != "" {
		pages, err = ParsePages(matches[5])
		if err != nil {
			return nil, notes
		}

		newNotes += fmt.Sprintf(", p.%s", pages.String())
	}

	if matches[7] != "" && pages != nil {
		recipeNumber, err = strconv.ParseUint(matches[7], 10, 64)
		if err != nil {
			return nil, notes
		}

		newNotes += fmt.Sprintf(", %s", ordinal(recipeNumber))
	}

	newNotes += "_"

	book := &Book{
		ISBN13:       isbn13,
		Pages:        pages,
		RecipeNumber: uint(recipeNumber),
	}

	return book, newNotes
}

func ordinal(n uint64) string {
	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}
