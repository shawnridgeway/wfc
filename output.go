package wfc

import (
	"image"
	"image/color"
)

// General struct for common functionality between CompleteImage and IncompleteImage
type GeneratedImage struct {
	data [][]color.Color // data to render
}

func (gi GeneratedImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (gi GeneratedImage) Bounds() image.Rectangle {
	if len(gi.data) < 1 {
		return image.Rect(0, 0, 0, 0)
	}
	return image.Rect(0, 0, len(gi.data[0]), len(gi.data))
}

func (gi GeneratedImage) At(x, y int) color.Color {
	return gi.data[x][y]
}
