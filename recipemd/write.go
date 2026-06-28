package recipemd

import (
	"fmt"
	"io"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
)

// writeMarkdown emits an interchange recipe as a RecipeMD document. It is the
// inverse of parseRecipe and round-trips stably for recipes that originated as
// RecipeMD. Ingredient amounts are italicised in RecipeMD's recommended style.
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
	writeIngredients(&b, ir.Ingredients)

	b.WriteString("\n---\n")
	writeSections(&b, ir.Instructions)

	_, err := io.WriteString(w, b.String())
	return err
}

// writeSections writes titled instruction lists: a `## Title` header for named
// sections, then the steps as a numbered list.
func writeSections(b *strings.Builder, sections []formats.TitledList) {
	for _, sec := range sections {
		if sec.Title != "" {
			fmt.Fprintf(b, "\n## %s\n", sec.Title)
		}
		b.WriteByte('\n')
		for i, item := range sec.List {
			fmt.Fprintf(b, "%d. %s\n", i+1, item)
		}
	}
}

// writeIngredients writes ingredient groups as bullet lists, italicising each
// item's amount/unit.
func writeIngredients(b *strings.Builder, groups []formats.IngredientGroup) {
	for _, g := range groups {
		if g.Title != "" {
			fmt.Fprintf(b, "\n## %s\n", g.Title)
		}
		b.WriteByte('\n')
		for _, iu := range g.Items {
			fmt.Fprintf(b, "- %s\n", formatIngredient(iu))
		}
	}
}

// formatIngredient renders one ingredient as a RecipeMD bullet body, with the
// amount and unit italicised and any note in parentheses.
func formatIngredient(iu ingredients.IngredientUse) string {
	var parts []string
	if qty := ingredients.RenderQuantity(iu); qty != "" {
		parts = append(parts, "*"+qty+"*")
	}
	parts = append(parts, iu.Ingredient.Name)
	line := strings.Join(parts, " ")
	if iu.Note != "" {
		line += " (" + iu.Note + ")"
	}
	return line
}
