package protected

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"unicode"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/utils"
)

const (
	// defaultQuestionCount is how many questions a protected archive carries.
	defaultQuestionCount = 8
	// defaultRequiredCorrect is how many of those questions must be answered
	// correctly to derive the decryption password.
	defaultRequiredCorrect = 4
	// minWordLength avoids picking trivially short words as answers.
	minWordLength = 4
)

// questionMaker functions return a question and its answer from a recipe, given
// the locator that tells the reader which recipe to consult. ok is false if this
// recipe can't furnish a question of this kind.
type questionMaker func(ir formats.InterchangeRecipe, locator string) (question, answer string, ok bool)

var (
	questionMakers = []questionMaker{
		questionRecipeTitle,
		questionRecipeDescription,
		questionRecipeInstructions,
	}
	sentenceSplit = regexp.MustCompile(`\s+`)
)

// prepareQuestions generates up to count questions (and their answers) spread
// across the provided recipes, using at most one question per recipe.
func prepareQuestions(recipes []formats.InterchangeRecipe, count int) (questions, answers []string, err error) {
	for _, rID := range rand.Perm(len(recipes)) {
		ir := recipes[rID]
		locator := recipeLocator(ir)

		for _, qmID := range rand.Perm(len(questionMakers)) {
			q, a, ok := questionMakers[qmID](ir, locator)
			if !ok {
				continue
			}
			// The locator names the recipe (eg. by title); if the answer is part
			// of it the question gives itself away, so skip that pairing and try
			// another kind of question.
			if locatorRevealsAnswer(locator, a) {
				continue
			}

			questions = append(questions, q)
			answers = append(answers, a)
			break
		}

		if len(questions) >= count {
			break
		}
	}

	if len(questions) < count {
		return nil, nil, fmt.Errorf("there are too few suitable recipes to generate %d questions (got %d)", count, len(questions))
	}

	return questions, answers, nil
}

// locatorRevealsAnswer reports whether the answer is already contained in the
// locator text, which would let the reader answer without consulting the recipe.
func locatorRevealsAnswer(locator, answer string) bool {
	a := normalize(answer)
	return a != "" && strings.Contains(normalize(locator), a)
}

func questionRecipeTitle(ir formats.InterchangeRecipe, locator string) (string, string, bool) {
	// Brackets suggest a subtitle, which may be hard to read back exactly — skip.
	if ir.Title == "" || strings.Contains(ir.Title, "(") || strings.Contains(ir.Title, "[") {
		return "", "", false
	}

	return locator + "What is the recipe's title?", ir.Title, true
}

func questionRecipeDescription(ir formats.InterchangeRecipe, locator string) (string, string, bool) {
	if ir.Description == "" {
		return "", "", false
	}

	sentences := strings.Split(ir.Description, ". ")
	sentenceLoc, sentence := pickNamedListItem(sentences)
	wordLoc, word := pickNamedListItem(sentenceSplit.Split(strings.TrimSpace(sentence), -1))
	word = trimPunc(word)
	if word == "" {
		return "", "", false
	}

	question := locator + fmt.Sprintf(
		"In the recipe's description, what is the %s word of the %s sentence?",
		wordLoc, sentenceLoc,
	)
	return question, word, true
}

func questionRecipeInstructions(ir formats.InterchangeRecipe, locator string) (string, string, bool) {
	if len(ir.Instructions) == 0 {
		return "", "", false
	}

	section := ir.Instructions[rand.Intn(len(ir.Instructions))]
	if len(section.List) == 0 {
		return "", "", false
	}

	var secLocator string
	if section.Title == "" {
		secLocator = "In the recipe's instructions"
	} else {
		secLocator = fmt.Sprintf("Within the part of the recipe's instructions labeled '%s'", section.Title)
	}

	lineLoc, line := pickNamedListItem(section.List)
	wordLoc, word := pickNamedListItem(sentenceSplit.Split(strings.TrimSpace(line), -1))
	word = trimPunc(word)
	if word == "" {
		return "", "", false
	}

	// Phrasing improvement in English.
	if lineLoc == "last" {
		lineLoc = "final"
	}

	question := locator + fmt.Sprintf(
		"%s, what is the %s word of the %s step?",
		secLocator, wordLoc, lineLoc,
	)
	return question, word, true
}

// recipeLocator produces a human-readable instruction for finding the recipe.
// When the recipe's ID encodes a book reference (an ISBN URN) the physical page is
// used; otherwise it falls back to the recipe's title (in which case the title
// itself must not become a required answer — see locatorRevealsAnswer).
func recipeLocator(ir formats.InterchangeRecipe) string {
	br, err := utils.NewBookRefFromURN(ir.ID)
	if err != nil || len(br.Pages) == 0 {
		if ir.Title != "" {
			return fmt.Sprintf("Look at the recipe titled '%s'. ", ir.Title)
		}
		return ""
	}

	pageOrdinal := ""
	if br.RecipeNumber > 0 {
		pageOrdinal = " " + utils.Ordinal(uint64(br.RecipeNumber), true)
	}

	return fmt.Sprintf("Look at the%s recipe on page %s. ", pageOrdinal, br.Pages.First())
}

// pickNamedListItem returns a human locator and an item from a list, eg. from
// ["a", "b", "c"] it might return ("second", "b").
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

// trimPunc trims punctuation from either end of a string.
func trimPunc(str string) string {
	return strings.TrimFunc(str, unicode.IsPunct)
}
