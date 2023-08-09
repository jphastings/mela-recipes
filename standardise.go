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
)

var kebabCaser = regexp.MustCompile(`[^a-z0-9]+`)

func (r *Recipe) Standardize(network bool) error {
	r.Filename = kebabCaser.ReplaceAllString(strings.ToLower(r.Title), "-")

	if err := bookFromNotes(r); err != nil {
		return err
	}

	for _, i := range r.Images {
		if err := i.Optimize(); err != nil {
			return err
		}
	}

	if network {
		if err := linkFromOpenLibrary(r); err != nil {
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
	if r.Book() == nil {
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
