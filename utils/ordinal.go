package utils

import "fmt"

// Ordinal renders n as an English ordinal. With useWords, the first three are
// spelled out ("first", "second", "third"); otherwise (and for all larger
// numbers) the numeric form is used ("1st", "2nd", "3rd", "4th", "11th", ...).
func Ordinal(n uint64, useWords bool) string {
	if useWords {
		switch n {
		case 1:
			return "first"
		case 2:
			return "second"
		case 3:
			return "third"
		}
	}

	if (n%100)/10 == 1 {
		return fmt.Sprintf("%dth", n)
	}
	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}
