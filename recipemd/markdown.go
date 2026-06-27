package recipemd

import (
	"errors"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// errNotRecipeMD marks a Markdown document that doesn't follow the RecipeMD
// structure (an H1 title and two `---` rules). It is returned so the .md file can
// be skipped with a clear message rather than producing an empty recipe.
var errNotRecipeMD = errors.New("not a RecipeMD document (needs an H1 title and two '---' rules)")

// parseRecipe maps a RecipeMD Markdown document into the interchange format.
//
// A RecipeMD document is three zones separated by the first two thematic breaks:
// a head (title, description, tags, yield), an ingredient list, and free-form
// instructions.
func parseRecipe(source []byte) (formats.InterchangeRecipe, error) {
	doc := goldmark.DefaultParser().Parse(text.NewReader(source))

	var blocks []ast.Node
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		blocks = append(blocks, c)
	}

	var breaks []int
	for i, b := range blocks {
		if _, ok := b.(*ast.ThematicBreak); ok {
			breaks = append(breaks, i)
			if len(breaks) == 2 {
				break
			}
		}
	}
	if len(breaks) < 2 {
		return formats.InterchangeRecipe{}, errNotRecipeMD
	}

	ir := formats.NewInterchangeRecipe()
	parseHead(blocks[:breaks[0]], source, &ir)
	ir.Ingredients = parseGroupedList(blocks[breaks[0]+1:breaks[1]], source, true)
	ir.Instructions = parseGroupedList(blocks[breaks[1]+1:], source, false)

	if ir.Title == "" {
		return formats.InterchangeRecipe{}, errNotRecipeMD
	}
	return ir, nil
}

// parseHead reads the title (first H1), description paragraphs, and the tag /
// yield paragraphs (a paragraph that is entirely italic / bold respectively).
func parseHead(blocks []ast.Node, source []byte, ir *formats.InterchangeRecipe) {
	var description []string
	for _, b := range blocks {
		switch n := b.(type) {
		case *ast.Heading:
			if n.Level == 1 && ir.Title == "" {
				ir.Title = strings.TrimSpace(nodeText(n, source))
			}
		case *ast.Paragraph:
			if tags, yield, ok := metadataParagraph(n, source); ok {
				ir.Tags = append(ir.Tags, tags...)
				if yield != "" {
					ir.Yield = yield
				}
			} else if para := strings.TrimSpace(nodeText(n, source)); para != "" {
				description = append(description, para)
			}
		}
	}
	ir.Description = strings.Join(description, "\n\n")
}

// metadataParagraph recognises RecipeMD's tag and yield paragraph(s). A paragraph
// counts as metadata when every span it contains is emphasis (italic → tags,
// bold → yield) — they may share one paragraph on separate lines. A paragraph
// with any plain prose is description and returns ok=false.
func metadataParagraph(p *ast.Paragraph, source []byte) (tags []string, yield string, ok bool) {
	for c := p.FirstChild(); c != nil; c = c.NextSibling() {
		switch n := c.(type) {
		case *ast.Emphasis:
			txt := strings.TrimSpace(nodeText(n, source))
			switch n.Level {
			case 1:
				tags = append(tags, splitList(txt)...)
			case 2:
				yield = txt
			}
		case *ast.Text:
			// Allow the soft/hard line break between an italic and a bold line;
			// any real text means this is a description paragraph.
			if strings.TrimSpace(string(n.Segment.Value(source))) != "" {
				return nil, "", false
			}
		default:
			return nil, "", false
		}
	}
	return tags, yield, len(tags) > 0 || yield != ""
}

// parseGroupedList turns the blocks of an ingredient or instruction zone into
// titled sections. `##`/`###` headers start a new section; ungrouped items land
// in a leading untitled section. Ingredient items come from list items only;
// instruction steps come from list items and paragraphs.
func parseGroupedList(blocks []ast.Node, source []byte, ingredients bool) []formats.TitledList {
	var sections []formats.TitledList
	cur := formats.TitledList{}
	flush := func() {
		if len(cur.List) > 0 {
			sections = append(sections, cur)
		}
	}

	for _, b := range blocks {
		switch n := b.(type) {
		case *ast.Heading:
			flush()
			cur = formats.TitledList{Title: strings.TrimSpace(nodeText(n, source))}
		case *ast.List:
			for li := n.FirstChild(); li != nil; li = li.NextSibling() {
				if item := strings.TrimSpace(nodeText(li, source)); item != "" {
					cur.List = append(cur.List, item)
				}
			}
		case *ast.Paragraph:
			if !ingredients {
				if step := strings.TrimSpace(nodeText(n, source)); step != "" {
					cur.List = append(cur.List, step)
				}
			}
		}
	}
	flush()
	return sections
}

// splitList splits a comma-separated string into trimmed, non-empty parts.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// nodeText extracts the plain text of a node and its descendants. Emphasis markers
// are dropped (so `*2 g* salt` becomes `2 g salt`); soft breaks become spaces.
func nodeText(n ast.Node, source []byte) string {
	var sb strings.Builder
	_ = ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := node.(type) {
		case *ast.Text:
			sb.Write(t.Segment.Value(source))
			if t.SoftLineBreak() {
				sb.WriteByte(' ')
			} else if t.HardLineBreak() {
				sb.WriteByte('\n')
			}
		case *ast.String:
			sb.Write(t.Value)
		}
		return ast.WalkContinue, nil
	})
	return sb.String()
}
