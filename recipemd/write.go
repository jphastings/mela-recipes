package recipemd

import (
	"fmt"
	"io"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
)

// writeMarkdown emits an interchange recipe as a RecipeMD document. It is the
// inverse of parseRecipe and round-trips stably for recipes that originated as
// RecipeMD. Amounts can't be re-italicised (the interchange holds each ingredient
// as a single string), so items are written plain.
func writeMarkdown(w io.Writer, ir formats.InterchangeRecipe) error {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n", ir.Title)

	if ir.Description != "" {
		fmt.Fprintf(&b, "\n%s\n", ir.Description)
	}
	if len(ir.Tags) > 0 {
		fmt.Fprintf(&b, "\n*%s*\n", strings.Join(ir.Tags, ", "))
	}
	if ir.Yield != "" {
		fmt.Fprintf(&b, "\n**%s**\n", ir.Yield)
	}

	b.WriteString("\n---\n")
	writeSections(&b, ir.Ingredients, false)

	b.WriteString("\n---\n")
	writeSections(&b, ir.Instructions, true)

	_, err := io.WriteString(w, b.String())
	return err
}

// writeSections writes titled lists: a `## Title` header for named sections, then
// the items as a bullet list (ingredients) or a numbered list (instructions).
func writeSections(b *strings.Builder, sections []formats.TitledList, numbered bool) {
	for _, sec := range sections {
		if sec.Title != "" {
			fmt.Fprintf(b, "\n## %s\n", sec.Title)
		}
		b.WriteByte('\n')
		for i, item := range sec.List {
			if numbered {
				fmt.Fprintf(b, "%d. %s\n", i+1, item)
			} else {
				fmt.Fprintf(b, "- %s\n", item)
			}
		}
	}
}
