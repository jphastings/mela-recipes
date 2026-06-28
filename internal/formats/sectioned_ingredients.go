package formats

import "github.com/jphastings/recipes/internal/ingredients"

// SectionedToIngredientGroups parses a sectioned ingredient sequence (the string
// form Mela and Paprika store natively) into structured ingredient groups,
// parsing each line via the ingredients helper.
func SectionedToIngredientGroups(ss SectionedSequence) []IngredientGroup {
	tls := SectionedToTitledLists(ss)
	groups := make([]IngredientGroup, 0, len(tls))
	for _, tl := range tls {
		g := IngredientGroup{Title: tl.Title}
		for i, line := range tl.List {
			g.Items = append(g.Items, ingredients.ParseOrItem(line, i))
		}
		groups = append(groups, g)
	}
	return groups
}

// IngredientsToSectioned renders structured ingredient groups back to the string
// sectioned form, rendering each item with FormatIngredientUse.
func IngredientsToSectioned(groups []IngredientGroup) SectionedSequence {
	tls := make([]TitledList, 0, len(groups))
	for _, g := range groups {
		tl := TitledList{Title: g.Title}
		for _, iu := range g.Items {
			tl.List = append(tl.List, ingredients.FormatIngredientUse(iu))
		}
		tls = append(tls, tl)
	}
	return TitledListsToSectioned(tls)
}
