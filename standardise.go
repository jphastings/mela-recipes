package mela

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var kebabCaser = regexp.MustCompile(`[^a-z0-9]+`)

func (r *Recipe) Standardize() error {
	r.Filename = kebabCaser.ReplaceAllString(strings.ToLower(r.Title), "-")

	if err := bookFromNotes(r); err != nil {
		return err
	}

	for _, i := range r.Images {
		if err := i.Optimize(); err != nil {
			return err
		}
	}

	return nil
}

var extractor = regexp.MustCompile(`(?i)(\s*)((?:isbn:? ?|_)([0-9X-]+)\r?\n?((?:, p.|pages?:? ?)([^_\s,]+)\r?\n?((?:recipe:? ?|, )?(\d+)(?:[a-z]{2})?\r?\n?)?)?)_?(\s*)`)

func bookFromNotes(r *Recipe) error {
	matches := extractor.FindStringSubmatch(r.Notes)
	if matches == nil {
		return nil
	}

	var newNotes string
	around := strings.SplitN(r.Notes, matches[0], 2)
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
		return err
	}

	newNotes += fmt.Sprintf("_%s", isbn13)

	var pages Pages
	var recipeNumber uint64

	if matches[5] != "" {
		pages, err = ParsePages(matches[5])
		if err != nil {
			return err
		}

		newNotes += fmt.Sprintf(", p.%s", pages.String())
	}

	if matches[7] != "" && pages != nil {
		recipeNumber, err = strconv.ParseUint(matches[7], 10, 64)
		if err != nil {
			return err
		}

		newNotes += fmt.Sprintf(", %s", ordinal(recipeNumber))
	}

	newNotes += "_"

	if err := r.SetBook(isbn13, pages, uint(recipeNumber)); err != nil {
		return err
	}
	r.Notes = newNotes

	return nil
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
