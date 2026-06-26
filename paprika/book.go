package paprika

import "github.com/jphastings/recipes/utils"

// Book returns the physical book reference encoded in the recipe's uid (as an
// ISBN URN), or the zero value if the uid is not such a reference.
func (r *Recipe) Book() utils.BookRef {
	book, err := utils.NewBookRefFromURN(r.UID)
	if err != nil {
		return utils.BookRef{}
	}
	return book
}
