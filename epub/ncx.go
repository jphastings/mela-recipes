package epub

import (
	"encoding/xml"
	"io"
)

type NCX struct {
	XMLName   xml.Name  `xml:"ncx"`
	Version   string    `xml:"version,attr"`
	DocTitle  DocTitle  `xml:"docTitle"`
	DocAuthor DocAuthor `xml:"docAuthor"`
	NavMap    NavMap    `xml:"navMap"`
}

type DocTitle struct {
	Text string `xml:"text"`
}

type DocAuthor struct {
	Text string `xml:"text"`
}

type NavMap struct {
	NavPoints []NavPoint `xml:"navPoint"`
}

type NavPoint struct {
	ID        string     `xml:"id,attr"`
	PlayOrder string     `xml:"playOrder,attr"`
	Label     NavLabel   `xml:"navLabel"`
	Content   NavContent `xml:"content"`
	Children  []NavPoint `xml:"navPoint"` // Supports nested navPoints.
}

type NavLabel struct {
	Text string `xml:"text"`
}

type NavContent struct {
	Src string `xml:"src,attr"`
}

func parseNCX(r io.Reader) (NCX, error) {
	var ncx NCX
	err := xml.NewDecoder(r).Decode(&ncx)
	if err != nil {
		return NCX{}, err
	}
	return ncx, nil
}
