package mela

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"unicode"
)

// questionMaker functions return a question and answer of a particular type from a recipe
type questionMaker func(*Recipe) (string, string, bool)

var (
	// questionMakers holds the set of question makers that will be used to generate questions for ownership proof
	questionMakers = []questionMaker{questionRecipeTitle, questionRecipeDescription, questionRecipeInstructions}
	// sentenceSplit is the regular expression used to split a sentence into words
	sentenceSplit = regexp.MustCompile(`\s+`)
)

//--- Generating questions ---

func (pr *ProtectedRecipes) PrepareOwnershipQuestions() ([]string, []string, error) {
	pr.questions = []string{}
	pr.answers = []string{}
	rIDs := rand.Perm(len(pr.unprotectedRecipes))

	for _, rID := range rIDs {
		for _, qmID := range rand.Perm(len(questionMakers)) {
			q, a, ok := questionMakers[qmID](pr.unprotectedRecipes[rID])
			if !ok {
				continue
			}

			pr.questions = append(pr.questions, q)
			pr.answers = append(pr.answers, a)
			break
		}

		if len(pr.questions) >= questionCount {
			break
		}
	}

	if len(pr.questions) < questionCount {
		return nil, nil, fmt.Errorf("there are too few recipes to be able to generate %d questions", questionCount)
	}

	return pr.questions, pr.answers, nil
}

//--- Question Maker functions ---

// questionRecipeTitle returns a question asking for the title of a recipe, and the matching answer
func questionRecipeTitle(r *Recipe) (string, string, bool) {
	// Brackets means there's a subtitle which may be hard to interpret — skip it!
	if r.Title == "" || strings.Contains(r.Title, "(") || strings.Contains(r.Title, "[") {
		return "", "", false
	}

	question := recipeLocationText(r) + "What is the recipe's title?"
	answer := r.Title

	return question, answer, true
}

// questionRecipeDescription returns a question asking about the description of a recipe, and the matching answer
func questionRecipeDescription(r *Recipe) (string, string, bool) {
	if r.Text == "" {
		return "", "", false
	}

	sentences := strings.Split(r.Text, ". ")
	sentenceLoc, sentence := pickNamedListItem(sentences)
	wordLoc, word := pickNamedListItem(sentenceSplit.Split(strings.TrimSpace(sentence), -1))
	word = trimPunc(word)

	question := recipeLocationText(r) + fmt.Sprintf(
		"In the recipe's description, what is the %s word of the %s sentence?",
		wordLoc,
		sentenceLoc,
	)
	answer := word

	return question, answer, true
}

// questionRecipeInstructions returns a question asking about the instructions of a recipe, and the matching answer
func questionRecipeInstructions(r *Recipe) (string, string, bool) {
	if r.Instructions == "" {
		return "", "", false
	}

	lines := r.Instructions.Parse()
	secName, secLines := randMapItem(lines)
	if secLines == nil {
		return "", "", false
	}
	var secLocator string
	if secName == "" {
		secLocator = "In the recipe's instructions"
	} else {
		secLocator = fmt.Sprintf("Within the part of the recipe's instructions labeled '%s'", secName)
	}

	lineLoc, line := pickNamedListItem(secLines)
	wordLoc, word := pickNamedListItem(sentenceSplit.Split(strings.TrimSpace(line), -1))
	word = trimPunc(word)

	// Phrasing improvement in English
	if lineLoc == "last" {
		lineLoc = "final"
	}

	question := recipeLocationText(r) + fmt.Sprintf(
		"%s, what is the %s word of the %s step?",
		secLocator,
		wordLoc,
		lineLoc,
	)
	answer := word

	return question, answer, true
}

//--- Helpers ---

// recipeLocationText is a helper function that produces (English) instructions for humans on finding a given recipe in a book
func recipeLocationText(r *Recipe) string {
	pageOrdinal := ""
	if r.Book().RecipeNumber > 0 {
		pageOrdinal = " " + ordinal(uint64(r.Book().RecipeNumber), true)
	}

	return fmt.Sprintf(
		"Look at the%s recipe on page %s. ",
		pageOrdinal,
		r.Book().Pages.First(),
	)
}

// pickNamedListItem returns a human locator and an item from a list, eg. from ["a", "b", "c"] might return "second", "b"
func pickNamedListItem(list []string) (string, string) {
	var possItems [][]string
	if len(list) >= 1 && len(list[0]) >= minWordLength {
		possItems = append(possItems, []string{"first", list[0]})
	}
	if len(list) >= 2 && len(list[len(list)-1]) >= minWordLength {
		possItems = append(possItems, []string{"last", list[len(list)-1]})
	}
	if len(list) >= 3 && len(list[1]) >= minWordLength {
		possItems = append(possItems, []string{"second", list[1]})
	}
	if len(list) >= 4 && len(list[len(list)-2]) >= minWordLength {
		possItems = append(possItems, []string{"second-to-last", list[len(list)-2]})
	}
	if len(list) >= 5 && len(list[2]) >= minWordLength {
		possItems = append(possItems, []string{"third", list[2]})
	}

	if len(possItems) == 0 {
		return "", ""
	}

	picked := possItems[rand.Intn(len(possItems))]
	return picked[0], picked[1]
}

// trimPunc trims punctuation from either end of a string
func trimPunc(str string) string {
	return strings.TrimFunc(str, func(r rune) bool {
		return unicode.IsPunct(r)
	})
}

// randMapItem returns a random key and (matched) value from from the given map
func randMapItem[T comparable, U any](m map[T]U) (T, U) {
	picked := rand.Intn(len(m))

	i := 0
	for k, v := range m {
		if i == picked {
			return k, v
		}
		i++
	}
	// We should never get here
	panic("cannot select random map item from empty map")
}
