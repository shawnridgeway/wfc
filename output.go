package wavefunctioncollapse

import (
	"image"
	"image/color"
)

// General struct for common functionality between CompleteImage and IncompleteImage
type GeneratedImage struct {
	*OverlappingModel                 // Underlying model
	Output            [][]color.Color // Output to render
}

func (gi GeneratedImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (gi GeneratedImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, gi.Fmx, gi.Fmy)
}

func (gi GeneratedImage) At(x, y int) color.Color {
	return gi.Output[y][x]
}
