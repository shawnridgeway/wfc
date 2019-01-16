package wavefunctioncollapse

import (
	"image"
	"image/color"
)

// General struct for common functionality between CompleteImage and IncompleteImage
type GeneratedImage struct {
	*Model
	output [][]color.Color
}

func NewCompleteImage(model *Model) *GeneratedImage {
	output := make([][]color.Color, model.fmy)
	for i := range output {
		output[i] = make([]color.Color, model.fmx)
	}
	for y := 0; y < model.fmy; y++ {
		for x := 0; x < model.fmx; x++ {
			for t := 0; t < model.t; t++ {
				if val := model.wave[x][y][t]; val {
					return model.colors[model.patterns[t][0]]
				}
			}
		}
	}
	return &GeneratedImage{Model: model, output: output}
}

func NewIncompleteImage(model *Model) *GeneratedImage {
	output := make([][]color.Color, model.fmy)
	for i := range output {
		output[i] = make([]color.Color, model.fmx)
	}
	var contributorNumber, r, g, b, a uint32
	for y := 0; y < model.fmy; y++ {
		for x := 0; x < model.fmx; x++ {
			contributorNumber, r, g, b, a = 0, 0, 0, 0, 0
			for dy := 0; dy < model.n; dy++ {
				for dx := 0; dx < model.n; dx++ {
					sx := x - dx
					if sx < 0 {
						sx += model.fmx
					}

					sy := y - dy
					if sy < 0 {
						sy += model.fmy
					}

					if !model.periodic && (sx+model.n > model.fmx || sy+model.n > model.fmy) {
						continue
					}

					for t := 0; t < model.t; t++ {
						if model.wave[sx][sy][t] {
							contributorNumber++
							clr := model.colors[model.patterns[t][dx+dy*model.n]]
							r += uint32(clr.R)
							g += uint32(clr.G)
							b += uint32(clr.B)
							a += uint32(clr.A)
						}
					}
				}
			}

			uR := uint8((r / contributorNumber) >> 24)
			uG := uint8((g / contributorNumber) >> 24)
			uB := uint8((b / contributorNumber) >> 24)
			uA := uint8((a / contributorNumber) >> 24)
			output[y][x] = color.RGBA{uR, uG, uB, uA}
		}
	}
	return &GeneratedImage{Model: model, output: output}
}

func (gi *GeneratedImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (gi *GeneratedImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, gi.fmx, gi.fmy)
}

func (gi *GeneratedImage) At(x, y int) color.Color {
	return gi.output[y][x]
}
