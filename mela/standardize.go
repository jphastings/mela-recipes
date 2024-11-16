package mela

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/utils"
)

func (r *Recipe) Standardize() ([]standardize.Std, error) {
	var stds []standardize.Std
	var stdApplied bool
	if r.filename, stdApplied = standardize.Filename(r.filename, r.Title); stdApplied {
		stds = append(stds, standardize.StdFilename)
	}

	if stdApplied, err := bookFromNotes(r); err == nil {
		return nil, err
	} else if stdApplied {
		stds = append(stds, standardize.StdISBN)
	}

	if r.Images == nil {
		r.Images = make([]utils.B64Image, 0)
	}

	if r.Categories == nil {
		r.Categories = make([]string, 0)
	}

	var imagesResized bool
	for i, img := range r.Images {
		newImg, ok, err := img.Optimize()
		if err != nil {
			return nil, err
		}
		r.Images[i] = newImg
		imagesResized = imagesResized || ok
	}
	if imagesResized {
		stds = append(stds, standardize.StdImages)
	}

	// if useNetwork {
	// 	_ = linkFromOpenLibrary(r)
	// }

	return stds, nil
}

var extractor = regexp.MustCompile(`(?i)(\s*)((?:isbn:? ?|_)([0-9X-]+)\r?\n?((?:, p\.|pages?:? ?)([^_\s,]+)\r?\n?((?:recipe:? ?|, )?(\d+)(?:[a-z]{2})?\r?\n?)?)?)_?(\s*)`)

func bookFromNotes(r *Recipe) (bool, error) {
	matches := extractor.FindStringSubmatch(r.Notes)
	if matches == nil {
		return false, nil
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

	isbn13, err := utils.StandardizeISBN(matches[3])
	if err != nil {
		return false, err
	}

	newNotes += fmt.Sprintf("_%s", isbn13)

	var pages utils.Pages
	var recipeNumber uint64

	if matches[5] != "" {
		pages, err = utils.ParsePages(matches[5])
		if err != nil {
			return false, err
		}

		newNotes += fmt.Sprintf(", p.%s", pages.String())
	}

	if matches[7] != "" && pages != nil {
		recipeNumber, err = strconv.ParseUint(matches[7], 10, 64)
		if err != nil {
			return false, err
		}

		newNotes += fmt.Sprintf(", %s", ordinal(recipeNumber, false))
	}

	newNotes += "_"

	if err := r.SetBook(isbn13, pages, uint(recipeNumber)); err != nil {
		return false, err
	}
	r.Notes = newNotes

	return true, nil
}

func ordinal(n uint64, useWords bool) string {
	if useWords {
		switch n {
		case 1:
			return "first"
		case 2:
			return "second"
		case 3:
			return "third"
		}
	}

	if (n%100)/10 == 1 {
		return fmt.Sprintf("%dth", n)
	}
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

type thingsResponse struct {
	Status string   `json:"status"`
	Result []string `json:"result"`
}

type getResponse struct {
	Status string `json:"status"`
	Result struct {
		Title string `json:"title"`
	} `json:"result"`
}

func linkFromOpenLibrary(r *Recipe) error {
	if r.Book().ISBN13 == "" {
		return nil
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	query := map[string]string{
		"type":    "/type/edition",
		"isbn_13": r.Book().ISBN13,
	}
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return err
	}

	qv := url.Values{}
	qv.Set("query", string(queryJSON))

	queryURL := url.URL{
		Scheme:   "https",
		Host:     "openlibrary.org",
		Path:     "/api/things",
		RawQuery: qv.Encode(),
	}

	vRes, err := client.Get(queryURL.String())
	if err != nil {
		return err
	}

	vBody, err := io.ReadAll(vRes.Body)
	if err != nil {
		return fmt.Errorf("unable to read OpenLibrary response: %w", err)
	}

	if vRes.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from OpenLibrary: %d (%s)", vRes.StatusCode, vBody)
	}

	var things thingsResponse
	if err := json.Unmarshal(vBody, &things); err != nil {
		return fmt.Errorf("unable to parse OpenLibrary response: %w", err)
	}

	if things.Status != "ok" {
		return fmt.Errorf("response status from OpenLibrary not ok: %s", things.Status)
	}

	if len(things.Result) == 0 {
		return fmt.Errorf("no books found with this ISBN in the OpenLibrary")
	}

	gv := url.Values{}
	gv.Set("key", things.Result[0])

	getURL := url.URL{
		Scheme:   "https",
		Host:     "openlibrary.org",
		Path:     "/api/get",
		RawQuery: gv.Encode(),
	}

	gRes, err := client.Get(getURL.String())
	if err != nil {
		return err
	}

	gBody, err := io.ReadAll(gRes.Body)
	if err != nil {
		return fmt.Errorf("unable to read OpenLibrary response: %w", err)
	}

	if gRes.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from OpenLibrary: %d (%s)", gRes.StatusCode, gBody)
	}

	var get getResponse
	if err := json.Unmarshal(gBody, &get); err != nil {
		return fmt.Errorf("unable to parse OpenLibrary response: %w", err)
	}

	if get.Status != "ok" {
		return fmt.Errorf("response status from OpenLibrary not ok: %s", get.Status)
	}

	r.Link = get.Result.Title
	return nil
}
