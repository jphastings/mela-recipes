package mela

import (
	"fmt"
	"image"
	"strconv"
	"strings"
	"time"
)

type Recipe interface {
	String() string
	Book() *Book
	SetBook(isbn string, pages Pages, index uint) error

	ID() string
	Title() string
	Link() string
	Text() string
	Ingredients() []string
	Instructions() []string
	Nutrition() string
	Categories() []string
	Notes() string

	Images(func(image.Image, error))
	Yield() (uint64, error)
	PrepTime() (*time.Duration, error)
	CookTime() (*time.Duration, error)
	TotalTime() (*time.Duration, error)
}

type Book struct {
	ISBN13       string
	Pages        Pages
	RecipeNumber uint
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
	RawNotes        string   `json:"notes"`

	RawImages    []string `json:"images"`
	RawYield     string   `json:"yield"`
	RawPrepTime  string   `json:"prepTime"`
	RawCookTime  string   `json:"cookTime"`
	RawTotalTime string   `json:"totalTime"`
}

func (r *RawRecipe) String() string { return fmt.Sprintf("Recipe for: %s", r.RawTitle) }

func (r *RawRecipe) ID() string             { return r.RawID }
func (r *RawRecipe) Title() string          { return r.RawTitle }
func (r *RawRecipe) Link() string           { return r.RawLink }
func (r *RawRecipe) Text() string           { return r.RawText }
func (r *RawRecipe) Categories() []string   { return r.RawCategories }
func (r *RawRecipe) Ingredients() []string  { return strings.Split(r.RawIngredients, "\n") }
func (r *RawRecipe) Instructions() []string { return strings.Split(r.RawInstructions, "\n") }
func (r *RawRecipe) Nutrition() string      { return r.RawNutrition }
func (r *RawRecipe) Notes() string          { return r.RawNotes }

func (r *RawRecipe) Yield() (uint64, error)             { return strconv.ParseUint(r.RawYield, 10, 64) }
func (r *RawRecipe) PrepTime() (*time.Duration, error)  { return durationGuesser(r.RawPrepTime) }
func (r *RawRecipe) CookTime() (*time.Duration, error)  { return durationGuesser(r.RawCookTime) }
func (r *RawRecipe) TotalTime() (*time.Duration, error) { return durationGuesser(r.RawTotalTime) }
