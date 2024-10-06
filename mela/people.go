package mela

import "strconv"

type PeopleCount string

func (pc PeopleCount) Parse() (uint64, error) {
	return strconv.ParseUint(string(pc), 10, 64)
}
