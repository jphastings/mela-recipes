package mela

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
)

type Pages []PageRange
type PageRange []string

func ParsePages(pages string) (Pages, error) {
	var pageRanges []PageRange
	parts := strings.Split(pages, ",")
	for _, prs := range parts {
		ps := strings.Split(prs, "-")
		if len(ps) > 2 {
			return nil, fmt.Errorf("invalid page range: %s", prs)
		}
		pr := make(PageRange, len(ps))
		var err error
		for i, p := range ps {
			if p == "" {
				return nil, fmt.Errorf("invalid page number: %s", p)
			}
			if pr[i], err = url.QueryUnescape(p); err != nil {
				return nil, fmt.Errorf("invalid URL encoding: %s", p)
			}
		}
		pageRanges = append(pageRanges, pr)
	}

	return pageRanges, nil
}

func MustParsePages(str string) Pages {
	p, err := ParsePages(str)
	if err != nil {
		panic(err)
	}
	return p
}

func (p Pages) String() string {
	var parts []string
	for _, pr := range p {
		p := url.QueryEscape(pr[0])
		if len(pr) > 1 {
			p += "-" + url.QueryEscape(pr[1])
		}
		parts = append(parts, p)
	}

	return strings.Join(parts, ",")
}

// CorrectContractions returns a new Pages object replacing page spans like "145-6" with the more explicit "145-146"
func (p Pages) CorrectContractions() Pages {
	newP := make(Pages, len(p))

	for i, pr := range p {
		newP[i] = make([]string, len(pr))
		copy(newP[i], pr)

		if len(pr) != 2 {
			continue
		}

		a, err := strconv.Atoi(pr[0])
		if err != nil {
			continue
		}
		b, err := strconv.Atoi(pr[1])
		if err != nil {
			continue
		}
		if b >= a {
			continue
		}

		div := int(math.Pow(10, math.Ceil(math.Log10(float64(b)))))
		extra := a / div

		newP[i][1] = fmt.Sprintf("%d", extra*div+b)
	}

	return newP
}
