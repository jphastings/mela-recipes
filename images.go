package mela

import (
	"encoding/base64"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

func (r *RawRecipe) Images(onImage func(image.Image, error)) {
	for _, img64 := range r.RawImages {
		dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(img64))

		img, _, err := image.Decode(dec)
		onImage(img, err)
	}
}
