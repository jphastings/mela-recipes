package recipes

import (
	"github.com/jphastings/recipes/internal/formats"
)

func AvailableFormats() []formats.Format {
	return formats.AvailableFormats
}

func AvailableFormatsList() string {
	l := len(formats.AvailableFormats)
	str := ""
	for i, f := range formats.AvailableFormats {
		str += f.Name
		if i == l-1 {
			// No separator after the last
		} else if i == l-2 {
			if l > 2 {
				str += ","
			}
			str += " and "
		} else {
			str += ", "
		}
	}
	return str
}
