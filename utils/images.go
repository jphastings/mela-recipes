package utils

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"

	_ "golang.org/x/image/webp"

	"golang.org/x/image/draw"

	"github.com/gen2brain/jpegli"
)

// OptimizeImages optimises each image in the slice in place, reporting whether
// any image was actually resized or recompressed. Per-image failures are
// collected into the returned error and leave that image unchanged.
func OptimizeImages(images []B64Image) (bool, error) {
	var resized bool
	var errs []error
	for i, img := range images {
		newImg, ok, err := img.Optimize()
		if err != nil {
			errs = append(errs, fmt.Errorf("image %d: %w", i, err))
			continue
		}
		images[i] = newImg
		resized = resized || ok
	}
	return resized, errors.Join(errs...)
}

func FromFile(f fs.File) (B64Image, error) {
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty image file")
	}

	return data, nil
}

type B64Image []byte

func (i B64Image) Optimize() (B64Image, bool, error) {
	return i.OptimizeWithConfig(512, 512)
}

func (i B64Image) OptimizeWithConfig(maxWidth, maxHeight int) (B64Image, bool, error) {
	img, imgType, err := DecodeImage(bytes.NewReader(i))
	if err != nil {
		return i, false, err
	}

	var wasResized bool
	img, wasResized = resizeImage(img, maxWidth, maxHeight)
	if !wasResized && (imgType == "jpeg") {
		return i, false, nil
	}

	opts := jpegli.EncodingOptions{
		Quality:           75,
		FancyDownsampling: true,
	}

	buf := new(bytes.Buffer)
	if err := jpegli.Encode(buf, img, &opts); err != nil {
		return i, false, err
	}

	return buf.Bytes(), true, nil
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
