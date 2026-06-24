package induce

import (
	"path"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// maxImagePageDist bounds how far (in printed pages) an image outside every
// recipe's flow may still be claimed by the nearest recipe.
const maxImagePageDist = 2

type blockMeta struct {
	ord  int // reading-order index across the whole book
	page int // printed page in effect at this block (0 if unknown)
}

// pageInNode returns the printed-page number declared inside a block (e.g. an
// <a id="page_82">), or 0.
func pageInNode(n *html.Node) int {
	res := 0
	eachElement(n, func(e *html.Node) {
		for _, a := range e.Attr {
			if a.Key == "id" {
				if m := pageIDRe.FindStringSubmatch(a.Val); m != nil {
					if p, err := strconv.Atoi(m[1]); err == nil {
						res = p
					}
				}
			}
		}
	})
	return res
}

// indexBlocks numbers every leaf block across the book in reading order and
// tags it with the printed page in effect, so images and recipes can be located
// relative to one another.
func indexBlocks(docs []Document) (map[*html.Node]blockMeta, [][]*html.Node) {
	meta := map[*html.Node]blockMeta{}
	perDoc := make([][]*html.Node, len(docs))
	ord, page := 0, 0
	for i, d := range docs {
		lb := leafBlocks(d.Root)
		perDoc[i] = lb
		for _, b := range lb {
			if pg := pageInNode(b); pg > 0 {
				page = pg
			}
			meta[b] = blockMeta{ord: ord, page: page}
			ord++
		}
	}
	return meta, perDoc
}

// isDecorative reports whether an <img> is flagged as non-content via the ARIA
// conventions every accessible EPUB uses for ornaments and chapter art.
func isDecorative(e *html.Node) bool {
	for _, a := range e.Attr {
		if a.Key == "role" && strings.EqualFold(a.Val, "presentation") {
			return true
		}
		if a.Key == "aria-hidden" && strings.EqualFold(a.Val, "true") {
			return true
		}
	}
	return false
}

type imageRef struct {
	node *html.Node
	path string
	meta blockMeta
}

func collectImages(docs []Document, meta map[*html.Node]blockMeta, perDoc [][]*html.Node) []imageRef {
	var refs []imageRef
	for i, d := range docs {
		dir := path.Dir(d.Name)
		for _, b := range perDoc[i] {
			m := meta[b]
			eachElement(b, func(e *html.Node) {
				if e.Data != "img" || isDecorative(e) {
					return
				}
				for _, a := range e.Attr {
					if a.Key == "src" && a.Val != "" {
						refs = append(refs, imageRef{node: e, path: path.Clean(path.Join(dir, a.Val)), meta: m})
					}
				}
			})
		}
	}
	return refs
}

// domOwner returns the recipe an image belongs to by DOM containment: the
// tightest ancestor shared with any recipe. If that ancestor holds exactly one
// recipe (a per-recipe <section>/<div> wrapping the recipe and its photo), the
// answer is unambiguous; if it holds several (a flat layout under <body>), -1
// signals "use another signal".
func domOwner(img *html.Node, delimCount, delimUnit map[*html.Node]int) int {
	for a := img; a != nil; a = a.Parent {
		if c := delimCount[a]; c > 0 {
			if c == 1 {
				return delimUnit[a]
			}
			return -1
		}
	}
	return -1
}

type recipeSpan struct {
	headOrd, lastOrd int // reading-order range of the recipe's blocks
	minPage, maxPage int
	hasPage          bool
}

func minPageDist(page int, s recipeSpan) int {
	switch {
	case page < s.minPage:
		return s.minPage - page
	case page > s.maxPage:
		return page - s.maxPage
	default:
		return 0
	}
}

// assignOne picks the recipe an image belongs to. Printed pages are decisive
// where present: the owning recipe is the one whose page span contains the image
// or which it faces (a full-page photo on page N belongs to the recipe starting
// on page N+1), with reading-order proximity breaking ties — this is what makes
// a facing-page photo land on the right recipe regardless of which side, or
// whose markup it sits in. Failing that, flow containment, then nearest page.
func assignOne(ref imageRef, spans []recipeSpan) int {
	page, ord := ref.meta.page, ref.meta.ord
	if page > 0 {
		best, bestOrd := -1, 1<<30
		for i := range spans {
			s := spans[i]
			if !s.hasPage {
				continue
			}
			if (page >= s.minPage && page <= s.maxPage) || s.minPage == page+1 {
				if d := abs(s.headOrd - ord); d < bestOrd {
					bestOrd, best = d, i
				}
			}
		}
		if best >= 0 {
			return best
		}
	}
	for i := range spans { // no usable page match: trust content flow
		if ord >= spans[i].headOrd && ord <= spans[i].lastOrd {
			return i
		}
	}
	if page > 0 { // last resort: nearest recipe by page, bounded
		best, bestDist := -1, maxImagePageDist+1
		for i := range spans {
			if !spans[i].hasPage {
				continue
			}
			if d := minPageDist(page, spans[i]); d < bestDist {
				bestDist, best = d, i
			}
		}
		return best
	}
	return -1
}

// assignImages attaches each image to its owning recipe, dropping decorative
// chapter art and never double-counting.
func (p *Profile) assignImages(docs []Document, units []Unit, recipes []Recipe) {
	if len(recipes) == 0 {
		return
	}
	meta, perDoc := indexBlocks(docs)
	refs := collectImages(docs, meta, perDoc)

	// DOM containment: number every recipe delimiter's ancestor chain so an
	// image inside a per-recipe wrapper can be attributed directly.
	delimCount := map[*html.Node]int{}
	delimUnit := map[*html.Node]int{}
	for i, u := range units {
		for a := u.Delim; a != nil; a = a.Parent {
			delimCount[a]++
			delimUnit[a] = i
		}
	}

	spans := make([]recipeSpan, len(units))
	for i, u := range units {
		if len(u.Blocks) == 0 {
			continue
		}
		s := &spans[i]
		s.headOrd = meta[u.Blocks[0]].ord
		s.lastOrd = meta[u.Blocks[len(u.Blocks)-1]].ord
		for _, pg := range recipes[i].Pages {
			n, err := strconv.Atoi(pg)
			if err != nil {
				continue
			}
			if !s.hasPage || n < s.minPage {
				s.minPage = n
			}
			if !s.hasPage || n > s.maxPage {
				s.maxPage = n
			}
			s.hasPage = true
		}
	}

	type scored struct {
		path string
		ord  int
	}
	buckets := make([][]scored, len(recipes))
	for _, ref := range refs {
		best := domOwner(ref.node, delimCount, delimUnit) // strongest signal first
		if best < 0 {
			best = assignOne(ref, spans) // fall back to printed page / facing / flow
		}
		if best >= 0 {
			buckets[best] = append(buckets[best], scored{ref.path, ref.meta.ord})
		}
	}
	for i := range recipes {
		bs := buckets[i]
		sort.SliceStable(bs, func(a, b int) bool { return bs[a].ord < bs[b].ord })
		seen := map[string]bool{}
		for _, s := range bs {
			if !seen[s.path] {
				seen[s.path] = true
				recipes[i].Images = append(recipes[i].Images, s.path)
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
