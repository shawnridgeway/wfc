package testutils

import (
	"image"
	"image/png"
	"os"
)

func LoadImage(file string) (image.Image, error) {
	raw, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer raw.Close()
	img, _, err := image.Decode(raw)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func SaveImage(file string, img image.Image) error {
	// Check if file exists, remove if it does
	if _, err := os.Stat(file); err == nil {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	// Save new file
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	png.Encode(f, img)
	return nil
}

func CompareImages(source, target image.Image) bool {
	if target.Bounds().Max.X != source.Bounds().Max.X || target.Bounds().Max.Y != source.Bounds().Max.Y {
		return false
	}
	for x := 0; x < target.Bounds().Max.X; x++ {
		for y := 0; y < target.Bounds().Max.Y; y++ {
			if target.At(x, y) != source.At(x, y) {
				return false
			}
		}
	}
	return true
}
