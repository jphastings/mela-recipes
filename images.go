package mela

import (
	"github.com/h2non/bimg"
)

type ImageBytes []byte

func (i ImageBytes) Optimize() error {
	return i.OptimizeWithConfig(1024, 1024, []bimg.ImageType{bimg.WEBP})
}

func (i ImageBytes) OptimizeWithConfig(maxWidth, maxHeight int, fileTypes []bimg.ImageType) error {
	targetType, convert := i.shouldConvert(fileTypes)

	size, err := bimg.NewImage(i).Size()
	if err != nil {
		return err
	}

	newWidth, newHeight := resizeAspectRatio(size.Width, size.Height, maxWidth, maxHeight)
	if size.Width != newWidth || size.Height != newHeight {
		if i, err = bimg.NewImage(i).Resize(newWidth, newHeight); err != nil {
			return err
		}
		convert = true
	}

	if !convert {
		return nil
	}

	i, err = bimg.NewImage(i).Convert(targetType)
	return err
}

func (i ImageBytes) shouldConvert(acceptableTypes []bimg.ImageType) (bimg.ImageType, bool) {
	if len(acceptableTypes) == 0 {
		return bimg.WEBP, true
	}

	imType := bimg.DetermineImageType(i)
	convert := true
	for _, t := range acceptableTypes {
		if imType == t {
			convert = false
			break
		}
	}

	return acceptableTypes[0], convert
}

func resizeAspectRatio(width, height, maxWidth, maxHeight int) (int, int) {
	if width <= maxWidth && height <= maxHeight {
		return width, height
	}

	if width > maxWidth {
		height = height * maxWidth / width
		width = maxWidth
	}

	if height > maxHeight {
		width = width * maxHeight / height
		height = maxHeight
	}

	return width, height
}
