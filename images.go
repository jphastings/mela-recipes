package mela

import (
	"fmt"
)

type ImageBytes []byte

func (i ImageBytes) Optimize() error {
	return fmt.Errorf("not implemented")
}
