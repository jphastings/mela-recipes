package recipes_test

import (
	"testing"

	. "github.com/jphastings/recipes"
	"github.com/jphastings/recipes/cooklang"
	"github.com/jphastings/recipes/crouton"
	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/mela"
	"github.com/stretchr/testify/assert"
)

func TestParseDestination(t *testing.T) {
	tc := []struct {
		to               string
		overrideFilename string
		asType           AsType
		format           *formats.Format
	}{
		{"Mela", "", AsTypeAny, mela.FormatInfo},
		{"mela", "", AsTypeAny, mela.FormatInfo},
		{".melarecipe", "", AsTypeRecipe, mela.FormatInfo},
		{".melarecipes", "", AsTypeCollection, mela.FormatInfo},
		{"something.melarecipe", "something", AsTypeRecipe, mela.FormatInfo},
		{"another.melarecipes", "another", AsTypeCollection, mela.FormatInfo},

		{".protectedrecipes", "", AsTypeCollection, mela.ProtectedFormatInfo},
		{"book.protectedrecipes", "book", AsTypeCollection, mela.ProtectedFormatInfo},

		{"Crouton", "", AsTypeAny, crouton.FormatInfo},
		{"crouton", "", AsTypeAny, crouton.FormatInfo},
		{".crumb", "", AsTypeRecipe, crouton.FormatInfo},
		{"something.crumb", "something", AsTypeRecipe, crouton.FormatInfo},

		{"Cooklang", "", AsTypeAny, cooklang.FormatInfo},
		{"cooklang", "", AsTypeAny, cooklang.FormatInfo},
		{".cook", "", AsTypeRecipe, cooklang.FormatInfo},
		{"something.cook", "something", AsTypeRecipe, cooklang.FormatInfo},

		{"nope", "", AsTypeAny, nil},
		{".nope", "", AsTypeAny, nil},
		{"whatever.nope", "", AsTypeAny, nil},
	}

	for _, c := range tc {
		t.Run(c.to, func(t *testing.T) {
			overrideFilename, asType, format := ParseDestination(c.to)
			assert.Equal(t, c.overrideFilename, overrideFilename)
			assert.Equal(t, c.asType, asType)
			assert.Equal(t, c.format, format)
		})
	}
}
