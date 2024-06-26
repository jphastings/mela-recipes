package mela

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"golang.org/x/image/draw"

	"github.com/gen2brain/jpegli"
)

type B64Image []byte

func (i B64Image) Optimize() (B64Image, error) {
	return i.OptimizeWithConfig(512, 512)
}

func (i B64Image) OptimizeWithConfig(maxWidth, maxHeight int) (B64Image, error) {
	img, imgType, err := image.Decode(bytes.NewReader(i))
	if err != nil {
		return i, err
	}

	var wasResized bool
	img, wasResized = resizeImage(img, maxWidth, maxHeight)
	if !wasResized && (imgType == "jpeg") {
		return i, nil
	}

	opts := jpegli.EncodingOptions{
		Quality:           75,
		FancyDownsampling: true,
	}

	buf := new(bytes.Buffer)
	if err := jpegli.Encode(buf, img, &opts); err != nil {
		return i, err
	}

	return buf.Bytes(), nil
}

func resizeImage(src image.Image, maxWidth, maxHeight int) (image.Image, bool) {
	newWidth, newHeight, needsResize := resizeAspectRatio(src.Bounds().Dx(), src.Bounds().Dy(), maxWidth, maxHeight)
	if !needsResize {
		return src, false
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	return dst, true
}

func resizeAspectRatio(width, height, maxWidth, maxHeight int) (int, int, bool) {
	if width <= maxWidth && height <= maxHeight {
		return width, height, false
	}

	if width > maxWidth {
		height = height * maxWidth / width
		width = maxWidth
	}

	if height > maxHeight {
		width = width * maxHeight / height
		height = maxHeight
	}

	return width, height, true
}
