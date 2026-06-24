package induce

import (
	"io"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"golang.org/x/text/unicode/norm"
)

// Document is one parsed content file from a book.
type Document struct {
	Name string
	Root *html.Node
}

func ParseDocument(name string, r io.Reader) (Document, error) {
	root, err := htmlquery.Parse(r)
	if err != nil {
		return Document{}, err
	}
	return Document{Name: name, Root: root}, nil
}

var blockTags = map[string]bool{
	"p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true,
	"h6": true, "div": true, "li": true, "dt": true, "dd": true, "blockquote": true,
}

func classOf(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "class" {
			return strings.TrimSpace(a.Val)
		}
	}
	return ""
}

func hasBlockDescendant(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if blockTags[c.Data] || hasBlockDescendant(c) {
			return true
		}
	}
	return false
}

// leafBlocks returns, in document order, the block-level elements that contain
// no nested block. These are the atomic "lines" of a recipe (a heading, an
// ingredient paragraph, a method step) and never overlap, so their text can be
// concatenated without double-counting.
func leafBlocks(root *html.Node) []*html.Node {
	var out []*html.Node
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			switch {
			case c.Type != html.ElementNode:
				walk(c)
			case blockTags[c.Data] && !hasBlockDescendant(c):
				out = append(out, c)
			default:
				walk(c)
			}
		}
	}
	walk(root)
	return out
}

func eachElement(root *html.Node, fn func(*html.Node)) {
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			fn(n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
}

func normalize(s string) string {
	s = norm.NFC.String(s)
	s = strings.ReplaceAll(s, " ", " ")
	return strings.Join(strings.Fields(s), " ")
}

// text returns the node's normalized inner text.
func text(n *html.Node) string { return normalize(htmlquery.InnerText(n)) }

// textExcluding returns normalized inner text, skipping subtrees matching strip.
func textExcluding(n *html.Node, strip []Sel) string {
	var b strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, s := range strip {
				if s.Matches(n) {
					return
				}
			}
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return normalize(b.String())
}

// Unit is one segmented recipe candidate: its leaf blocks plus the delimiter.
type Unit struct {
	Doc    string
	Delim  *html.Node
	Blocks []*html.Node
}

func segment(docs []Document, spec UnitSpec) []Unit {
	var units []Unit
	for _, d := range docs {
		if spec.Mode == ModeContainer {
			eachElement(d.Root, func(n *html.Node) {
				if spec.Sel.Matches(n) {
					units = append(units, Unit{Doc: d.Name, Delim: n, Blocks: leafBlocks(n)})
				}
			})
			continue
		}
		// heading mode
		lb := leafBlocks(d.Root)
		var idxs []int
		for i, n := range lb {
			if spec.Sel.Matches(n) {
				idxs = append(idxs, i)
			}
		}
		for j, start := range idxs {
			end := len(lb)
			if j+1 < len(idxs) {
				end = idxs[j+1]
			}
			units = append(units, Unit{Doc: d.Name, Delim: lb[start], Blocks: lb[start:end]})
		}
	}
	return units
}

// source is the normalized concatenation of a unit's leaf-block text — the
// reference string the verification gate checks captured fields against.
func (u Unit) source() string {
	parts := make([]string, 0, len(u.Blocks))
	for _, b := range u.Blocks {
		parts = append(parts, htmlquery.InnerText(b))
	}
	return normalize(strings.Join(parts, " "))
}

// --- structural primitives (language-agnostic) ---
//
// Role detection does NOT match recipe vocabulary. It uses structure only:
// how long a line is, where it sits, how many siblings share its class,
// whether it leads with a numeral, and whether it is wholly emphasised. These
// generalise across languages and house styles; a model labeller can later
// implement the Labeler interface for books too irregular for structure alone.

func runeLen(s string) int { return len([]rune(s)) }

// numLeadRe matches a leading numeral in any script (Unicode Nd) or a common
// vulgar-fraction glyph — the cross-language signal that a line states a quantity.
var numLeadRe = regexp.MustCompile(`^[\p{Nd}\x{00BC}-\x{00BE}\x{2150}-\x{215E}]`)

// numAnyRe finds a numeral anywhere (used to tell a yield "Serves 4" from a
// heading "For the sauce" once emphasis has marked both as non-items).
var numAnyRe = regexp.MustCompile(`[\p{Nd}\x{00BC}-\x{00BE}\x{2150}-\x{215E}]`)

func startsNumeric(t string) bool { return numLeadRe.MatchString(t) }
func hasNumeral(t string) bool    { return numAnyRe.MatchString(t) }

// bareQtyRe matches a line that is *only* a quantity — numerals, fractions and
// separators, no letters. Hanging-indent layouts put such a line in its own
// block ("1", "½", "1–2") with the ingredient name in the next block.
var bareQtyRe = regexp.MustCompile(`^[\p{Nd}\x{00BC}-\x{00BE}\x{2150}-\x{215E}\s+×./,–\-]+$`)

func isBareQuantity(t string) bool { return t != "" && bareQtyRe.MatchString(t) }

var letterRe = regexp.MustCompile(`\p{L}`)

func hasLetters(t string) bool { return letterRe.MatchString(t) }

// firstNumber returns the first run of Unicode digits, e.g. "Serves 4" -> "4".
var digitsRe = regexp.MustCompile(`\p{Nd}+`)

func firstNumber(t string) string { return digitsRe.FindString(t) }

var emphTags = map[string]bool{"strong": true, "b": true, "em": true, "i": true}

// emphasised reports whether all of a block's visible text sits inside emphasis
// markup (<strong>/<b>/<em>/<i>). Cookbooks bold/italicise group headings and
// the serving line while leaving ingredient items plain — a reliable, wordless
// way to separate them.
func emphasised(n *html.Node) bool {
	var plain strings.Builder
	var any bool
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, inEmph bool) {
		if n.Type == html.TextNode {
			if strings.TrimSpace(n.Data) != "" {
				any = true
				if !inEmph {
					plain.WriteString(n.Data)
				}
			}
			return
		}
		e := inEmph || (n.Type == html.ElementNode && emphTags[n.Data])
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, e)
		}
	}
	walk(n, false)
	return any && strings.TrimSpace(plain.String()) == ""
}
