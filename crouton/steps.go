package crouton

import "github.com/jphastings/crouton-recipes/uuid"

type Steps []Step

type Step struct {
	Order     int       `json:"order"`
	IsSection bool      `json:"isSection"`
	Step      string    `json:"step"`
	UUID      uuid.UUID `json:"uuid"`
}
