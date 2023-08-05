package mela

import (
	"fmt"
	"net/url"
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
