package crouton

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
	"github.com/jphastings/recipes/internal/uuid"
)

// Sections in both ingredients and steps map to formats.TitledList groups, mirroring
// mela's sectioned sequences: items before the first heading live under Title "".
//
// Crouton stores sections as flat markers in the item list, so every item is
// associated with the most recent preceding heading. An interchange recipe whose
// unsectioned items follow a titled section therefore cannot be represented: on
// import those items are absorbed into the preceding section. Export always emits
// the unsectioned group first, so this only affects such hand-built interchange
// recipes (outside the Crouton-representable subset).

func ingredientsToGroups(ius []ingredients.IngredientUse) []formats.IngredientGroup {
	var groups []formats.IngredientGroup
	cur := formats.IngredientGroup{}
	for _, iu := range ius {
		if iu.Quantity.Type == ingredients.SectionMarker {
			groups = appendGroupIfNonEmpty(groups, cur)
			cur = formats.IngredientGroup{Title: iu.Ingredient.Name}
			continue
		}
		cur.Items = append(cur.Items, iu) // carried verbatim (UUIDs, note kept)
	}
	return appendGroupIfNonEmpty(groups, cur)
}

func appendGroupIfNonEmpty(groups []formats.IngredientGroup, g formats.IngredientGroup) []formats.IngredientGroup {
	if g.Title != "" || len(g.Items) > 0 {
		return append(groups, g)
	}
	return groups
}

func stepsToTitledLists(steps Steps) []formats.TitledList {
	var lists []formats.TitledList
	cur := formats.TitledList{}
	for _, s := range steps {
		if s.IsSection {
			lists = appendIfNonEmpty(lists, cur)
			cur = formats.TitledList{Title: s.Step}
			continue
		}
		cur.List = append(cur.List, s.Step)
	}
	return appendIfNonEmpty(lists, cur)
}

// appendIfNonEmpty keeps named sections (even when empty) but drops an empty
// leading group with no title, so the "" group only appears when it has items.
func appendIfNonEmpty(lists []formats.TitledList, tl formats.TitledList) []formats.TitledList {
	if tl.Title != "" || len(tl.List) > 0 {
		return append(lists, tl)
	}
	return lists
}

func groupsToIngredients(groups []formats.IngredientGroup) ([]ingredients.IngredientUse, error) {
	var out []ingredients.IngredientUse
	order := 0
	for _, g := range groups {
		if g.Title != "" {
			sec, err := ingredients.NewSection(g.Title, order)
			if err != nil {
				return nil, err
			}
			out = append(out, sec)
			order++
		}
		for _, iu := range g.Items {
			iu.Order = order // carried verbatim; only the order is re-derived
			out = append(out, iu)
			order++
		}
	}
	return out, nil
}

func titledListsToSteps(tls []formats.TitledList) (Steps, error) {
	var steps Steps
	order := 0
	appendStep := func(isSection bool, text string) error {
		u, err := uuid.NewUUID("")
		if err != nil {
			return err
		}
		steps = append(steps, Step{Order: order, IsSection: isSection, Step: text, UUID: u})
		order++
		return nil
	}
	for _, tl := range tls {
		if tl.Title != "" {
			if err := appendStep(true, tl.Title); err != nil {
				return nil, err
			}
		}
		for _, line := range tl.List {
			if err := appendStep(false, line); err != nil {
				return nil, err
			}
		}
	}
	return steps, nil
}

func minutesToPtr(m Minutes) *time.Duration {
	d := time.Duration(m)
	if d == 0 {
		return nil
	}
	return &d
}

func ptrToMinutes(d *time.Duration) Minutes {
	if d == nil {
		return 0
	}
	return Minutes(*d)
}

var leadingInt = regexp.MustCompile(`^\s*(\d+)`)

func parseYield(y string) PeopleCount {
	m := leadingInt.FindStringSubmatch(y)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return PeopleCount(n)
}

// isFreeFormYield reports whether y carries detail a crouton import (keeping only a
// positive whole-number serving count) would lose: anything that isn't exactly the
// positive integer it parses to, including "0" and non-numeric text, which both
// export back as an empty yield.
func isFreeFormYield(y string) bool {
	y = strings.TrimSpace(y)
	if y == "" {
		return false
	}
	n := int(parseYield(y))
	return n == 0 || strconv.Itoa(n) != y
}

