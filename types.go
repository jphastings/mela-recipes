package mela

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Recipe interface {
	String() string
	Book() *Book
	SetBook(isbn string, page uint) (Recipe, error)

	ID() string
	Title() string
	Link() string
	Text() string
	Ingredients() string
	Instructions() string
	Nutrition() string
	Categories() []string

	Images(func(image.Image, error))
	Yield() (uint64, error)
	PrepTime() (*time.Duration, error)
	CookTime() (*time.Duration, error)
	TotalTime() (*time.Duration, error)
}

type Book struct {
	ISBN13 string
	Page   uint
}

type RawRecipe struct {
	RawID           string   `json:"id"`
	RawTitle        string   `json:"title"`
	RawLink         string   `json:"link"`
	RawText         string   `json:"text"`
	RawIngredients  string   `json:"ingredients"`
	RawInstructions string   `json:"instructions"`
	RawNutrition    string   `json:"nutrition"`
	RawCategories   []string `json:"categories"`

	RawImages    []string `json:"images"`
	RawYield     string   `json:"yield"`
	RawPrepTime  string   `json:"prepTime"`
	RawCookTime  string   `json:"cookTime"`
	RawTotalTime string   `json:"totalTime"`
}

func (r RawRecipe) String() string { return fmt.Sprintf("Recipe for: %s", r.RawTitle) }

var isbnUrnParser = regexp.MustCompile("^urn:isbn:([0-9]{13})#page=([0-9]+)$")

func (r RawRecipe) Book() *Book {
	m := isbnUrnParser.FindStringSubmatch(r.RawID)
	fmt.Println(m)
	if len(m) == 0 {
		return nil
	}

	isbn13 := m[1]
	// Will always be a number, because regexp
	page, _ := strconv.Atoi(m[2])

	return &Book{
		ISBN13: isbn13,
		Page:   uint(page),
	}
}

func (r RawRecipe) SetBook(isbn10or13 string, page uint) (Recipe, error) {
	isbn13, err := validateISBN(isbn10or13)
	if err != nil {
		return nil, err
	}

	r.RawID = fmt.Sprintf("urn:isbn:%s#page=%d", isbn13, page)
	return r, nil
}

func (r RawRecipe) Images(onImage func(image.Image, error)) {
	for _, img64 := range r.RawImages {
		dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(img64))

		img, _, err := image.Decode(dec)
		onImage(img, err)
	}
}

func (r RawRecipe) ID() string           { return r.RawID }
func (r RawRecipe) Title() string        { return r.RawTitle }
func (r RawRecipe) Link() string         { return r.RawLink }
func (r RawRecipe) Text() string         { return r.RawText }
func (r RawRecipe) Categories() []string { return r.RawCategories }
func (r RawRecipe) Ingredients() string  { return r.RawIngredients }
func (r RawRecipe) Instructions() string { return r.RawInstructions }
func (r RawRecipe) Nutrition() string    { return r.RawNutrition }

func (r RawRecipe) Yield() (uint64, error)             { return strconv.ParseUint(r.RawYield, 10, 64) }
func (r RawRecipe) PrepTime() (*time.Duration, error)  { return durationGuesser(r.RawPrepTime) }
func (r RawRecipe) CookTime() (*time.Duration, error)  { return durationGuesser(r.RawCookTime) }
func (r RawRecipe) TotalTime() (*time.Duration, error) { return durationGuesser(r.RawTotalTime) }
