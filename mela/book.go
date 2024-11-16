package mela

import (
	"github.com/jphastings/recipes/utils"
)

func (r *Recipe) Book() utils.BookRef {
	book, err := utils.NewBookRefFromURN(r.ID)
	if err != nil {
		return utils.BookRef{}
	}
	return book
}

func (r *Recipe) SetBook(isbn10or13 string, pages utils.Pages, recipeNumber uint) error {
	isbn13, err := utils.StandardizeISBN(isbn10or13)
	if err != nil {
		return err
	}

	book := utils.BookRef{
		Book: utils.Book{
			ISBN13: isbn13,
		},
		Pages:        pages,
		RecipeNumber: recipeNumber,
	}
	r.ID = book.URN()

	return nil
}
