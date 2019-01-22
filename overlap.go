package wavefunctioncollapse

import (
	// "fmt"
	"image"
	"image/color"
	"math"
)

/**
 * OverlappingModel Type
 */
type OverlappingModel struct {
	*BaseModel                 // Underlying model of generic Wave Function Collapse algorithm
	N            int           // Size of patterns (ie pixel distance of influencing pixels)
	Colors       []color.Color // Array of unique colors in input
	Ground       int           // Id of the specific pattern to use as the bottom of the generation. A value of -1 means that this is unset
	Patterns     []Pattern     // Array of unique patterns in input
	Propagator   [][][][]int   // Table of which patterns (t2) mathch a given pattern (t1) at offset (dx, dy) [t1][dx][dy][t2]
	Fmxmn, Fmymn int           // Width and height of output, minus n
}

/**
 * Pattern Type
 */
type Pattern []int

/**
 * NewOverlappingModel
 * @param {image.Image} img The source image
 * @param {int} N Size of the patterns
 * @param {int} width The width of the generated image
 * @param {int} height The height of the generated image
 * @param {bool} periodicInput Whether the source image is to be considered as periodic / as a repeatable texture
 * @param {bool} periodicOutput Whether the generation should be periodic / a repeatable texture
 * @param {int} symmetry Allowed symmetries from 1 (no symmetry) to 8 (all mirrored / rotated variations)
 * @param {int} [ground=0] Id of the specific pattern to use as the bottom of the generation ( see https://github.com/mxgmn/WaveFunctionCollapse/issues/3#issuecomment-250995366 )
 * @returns *OverlappingModel A pointer to a new copy of the model
 */
func NewOverlappingModel(img image.Image, n, width, height int, periodicInput, periodicOutput bool, symmetry int, ground bool) *OverlappingModel {

	// Initialize model
	model := &OverlappingModel{BaseModel: &BaseModel{}}
	model.N = n
	model.Fmx = width
	model.Fmy = height
	model.Periodic = periodicOutput
	model.Ground = -1

	bounds := img.Bounds()
	dataWidth := bounds.Max.X
	dataHeight := bounds.Max.Y

	// Build up a palette of colors (by assigning numbers to unique color values)
	sample := make([][]int, dataWidth)
	for i := range sample {
		sample[i] = make([]int, dataHeight)
	}

	model.Colors = make([]color.Color, 0)
	colorMap := make(map[color.Color]int)

	for y := 0; y < dataHeight; y++ {
		for x := 0; x < dataWidth; x++ {
			color := img.At(x, y)
			if _, ok := colorMap[color]; !ok {
				colorMap[color] = len(model.Colors)
				model.Colors = append(model.Colors, color)
			}
			sample[x][y] = colorMap[color]
		}
	}

	// Extract various patterns from input (patterns are 1D arrays of sample codes)
	c := len(model.Colors)
	w := int(math.Pow(float64(c), float64(n*n)))

	// Given a transforming function, return a flattened array of the N*N pattern
	getPattern := func(transformer func(x, y int) int) Pattern {
		result := make(Pattern, n*n)
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				result[x+y*n] = transformer(x, y)
			}
		}
		return result
	}

	// Return a flattened array of the N*N pattern at (x, y) using sample codes
	patternFromSample := func(x, y int) Pattern {
		return getPattern(func(dx, dy int) int {
			return sample[(x+dx)%dataWidth][(y+dy)%dataHeight]
		})
	}

	rotate := func(p Pattern) Pattern {
		return getPattern(func(x, y int) int {
			return p[n-1-y+x*n]
		})
	}

	reflect := func(p Pattern) Pattern {
		return getPattern(func(x, y int) int {
			return p[n-1-x+y*n]
		})
	}

	// Compute a "hash" value for indexing patterns (unique for unique patterns)
	indexFromPattern := func(p Pattern) int {
		result := 0
		power := 1
		for i := 0; i < len(p); i++ {
			result += p[len(p)-1-i] * power
			power *= c
		}
		return result
	}

	// Reverse the hash of a pattern's index
	patternFromIndex := func(ind int) Pattern {
		residue := ind
		power := w
		result := make(Pattern, n*n)
		for i := 0; i < len(result); i++ {
			power /= c
			count := 0
			for residue >= power {
				residue -= power
				count++
			}
			result[i] = count
		}
		return result
	}

	// Build map of patterns (indexed by computed hash) to weights based on frequency in sample
	weights := make(map[int]int)
	weightsKeys := make([]int, 0)

	var horizontalBound, verticalBound int
	if periodicInput {
		horizontalBound = dataWidth
		verticalBound = dataHeight
	} else {
		horizontalBound = dataWidth - n + 1
		verticalBound = dataHeight - n + 1
	}
	for y := 0; y < verticalBound; y++ {
		for x := 0; x < horizontalBound; x++ {
			ps := make([]Pattern, 8, 8)
			ps[0] = patternFromSample(x, y)
			ps[1] = reflect(ps[0])
			ps[2] = rotate(ps[0])
			ps[3] = reflect(ps[2])
			ps[4] = rotate(ps[2])
			ps[5] = reflect(ps[4])
			ps[6] = rotate(ps[4])
			ps[7] = reflect(ps[6])
			for k := 0; k < symmetry; k++ {
				ind := indexFromPattern(ps[k])
				if _, ok := weights[ind]; ok {
					weights[ind]++
				} else {
					weightsKeys = append(weightsKeys, ind)
					weights[ind] = 1
				}
				if ground && y == verticalBound-1 && x == 0 && k == 0 {
					// Set groung pattern
					model.Ground = len(weightsKeys) - 1
				}
			}
		}
	}

	model.T = len(weightsKeys)

	// Store the patterns and cooresponding weights (stationary)
	model.Patterns = make([]Pattern, model.T)
	model.Stationary = make([]int, model.T)
	model.Propagator = make([][][][]int, model.T)
	for i, wk := range weightsKeys {
		model.Patterns[i] = patternFromIndex(wk)
		model.Stationary[i] = weights[wk]
	}

	// Initialize wave (to all true) and changes (to all false) fields
	model.Wave = make([][][]bool, model.Fmx)
	model.Changes = make([][]bool, model.Fmx)
	for x := 0; x < model.Fmx; x++ {
		model.Wave[x] = make([][]bool, model.Fmy)
		model.Changes[x] = make([]bool, model.Fmy)
		for y := 0; y < model.Fmy; y++ {
			model.Wave[x][y] = make([]bool, model.T)
			model.Changes[x][y] = false
			for t := 0; t < model.T; t++ {
				model.Wave[x][y][t] = true
			}
		}
	}

	// Check that the spaces n distance away have no conflicts
	agrees := func(p1, p2 Pattern, dx, dy int) bool {
		var xmin, xmax, ymin, ymax int

		if dx < 0 {
			xmin = 0
			xmax = dx + n
		} else {
			xmin = dx
			xmax = n
		}

		if dy < 0 {
			ymin = 0
			ymax = dy + n
		} else {
			ymin = dy
			ymax = n
		}

		for y := ymin; y < ymax; y++ {
			for x := xmin; x < xmax; x++ {
				if p1[x+n*y] != p2[x-dx+n*(y-dy)] {
					return false
				}
			}
		}

		return true
	}

	// Build table of which patterns can exist next to another
	for t := 0; t < model.T; t++ {
		model.Propagator[t] = make([][][]int, 2*n-1)
		for x := 0; x < 2*n-1; x++ {
			model.Propagator[t][x] = make([][]int, 2*n-1)
			for y := 0; y < 2*n-1; y++ {
				list := make([]int, 0)

				for t2 := 0; t2 < model.T; t2++ {
					if agrees(model.Patterns[t], model.Patterns[t2], x-n+1, y-n+1) {
						list = append(list, t2)
					}
				}

				model.Propagator[t][x][y] = make([]int, len(list))

				for k := 0; k < len(list); k++ {
					model.Propagator[t][x][y][k] = list[k]
				}
			}
		}
	}

	model.Fmxmn = model.Fmx - model.N
	model.Fmymn = model.Fmy - model.N

	return model
}

/**
 * OnBoundary
 */
func (model *OverlappingModel) OnBoundary(x, y int) bool {
	return !model.Periodic && (x > model.Fmxmn || y > model.Fmymn)
}

/**
 * Propagate
 * return: bool, change occured in this iteration
 */
func (model *OverlappingModel) Propagate() bool {
	change := false
	startLoop := -model.N + 1
	endLoop := model.N

	for x := 0; x < model.Fmx; x++ {
		for y := 0; y < model.Fmy; y++ {
			if model.Changes[x][y] {
				model.Changes[x][y] = false
				for dx := startLoop; dx < endLoop; dx++ {
					for dy := startLoop; dy < endLoop; dy++ {
						sx := x + dx
						sy := y + dy

						if sx < 0 {
							sx += model.Fmx
						} else if sx >= model.Fmx {
							sx -= model.Fmx
						}

						if sy < 0 {
							sy += model.Fmy
						} else if sy >= model.Fmy {
							sy -= model.Fmy
						}

						if !model.Periodic && (sx > model.Fmx || sy > model.Fmy) {
							continue
						}

						allowed := model.Wave[sx][sy]

						for t := 0; t < model.T; t++ {
							if !allowed[t] {
								continue
							}

							b := false
							prop := model.Propagator[t][model.N-1-dx][model.N-1-dy]
							for i := 0; i < len(prop) && !b; i++ {
								b = model.Wave[x][y][prop[i]]
							}

							if !b {
								model.Changes[sx][sy] = true
								change = true
								allowed[t] = false
							}
						}
					}
				}
			}
		}
	}

	return change
}

/**
 * Clear the internal state, then set ground pattern
 */
func (model *OverlappingModel) Clear() {
	model.ClearBase(model)
	if model.Ground != -1 && model.T > 1 {
		for x := 0; x < model.Fmx; x++ {
			for t := 0; t < model.T; t++ {
				if t != model.Ground {
					model.Wave[x][model.Fmy-1][t] = false
				}
			}

			model.Changes[x][model.Fmy-1] = true

			for y := 0; y < model.Fmy-1; y++ {
				model.Wave[x][y][model.Ground] = false
				model.Changes[x][y] = true
			}
		}

		for model.Propagate() {
			// Empty loop
		}
	}
}

/**
 * Create a GeneratedImage holding the data for a complete image
 */
func (model *OverlappingModel) RenderCompleteImage() GeneratedImage {
	output := make([][]color.Color, model.Fmy)
	for i := range output {
		output[i] = make([]color.Color, model.Fmx)
	}
	for y := 0; y < model.Fmy; y++ {
		for x := 0; x < model.Fmx; x++ {
			for t := 0; t < model.T; t++ {
				if val := model.Wave[x][y][t]; val {
					output[y][x] = model.Colors[model.Patterns[t][0]]
				}
			}
		}
	}
	return GeneratedImage{OverlappingModel: model, Output: output}
}

/**
 * Create a GeneratedImage holding the data for an incomplete image
 */
func (model *OverlappingModel) RenderIncompleteImage() GeneratedImage {
	output := make([][]color.Color, model.Fmy)
	for i := range output {
		output[i] = make([]color.Color, model.Fmx)
	}
	var contributorNumber, sR, sG, sB, sA uint32
	for y := 0; y < model.Fmy; y++ {
		for x := 0; x < model.Fmx; x++ {
			contributorNumber, sR, sG, sB, sA = 0, 0, 0, 0, 0

			for dy := 0; dy < model.N; dy++ {
				for dx := 0; dx < model.N; dx++ {
					sx := x - dx
					if sx < 0 {
						sx += model.Fmx
					}

					sy := y - dy
					if sy < 0 {
						sy += model.Fmy
					}

					if !model.Periodic && (sx > model.Fmxmn || sy > model.Fmymn) {
						continue
					}

					for t := 0; t < model.T; t++ {
						if model.Wave[sx][sy][t] {
							contributorNumber++
							r, g, b, a := model.Colors[model.Patterns[t][dx+dy*model.N]].RGBA()
							sR += r
							sG += g
							sB += b
							sA += a
						}
					}
				}
			}

			var uR, uG, uB, uA uint8
			if contributorNumber == 0 {
				uR = 127
				uG = 127
				uB = 127
				uA = 255
			} else {
				uR = uint8((sR / contributorNumber) >> 8)
				uG = uint8((sG / contributorNumber) >> 8)
				uB = uint8((sB / contributorNumber) >> 8)
				uA = uint8((sA / contributorNumber) >> 8)
			}

			output[y][x] = color.RGBA{uR, uG, uB, uA}
		}
	}
	return GeneratedImage{OverlappingModel: model, Output: output}
}

/**
 * Retrieve the RGBA data
 * returns: Image
 */
func (model *OverlappingModel) RenderImage() GeneratedImage {
	if model.IsGenerationSuccessful() {
		return model.RenderCompleteImage()
	} else {
		return model.RenderIncompleteImage()
	}
}

/**
 * Retrieve the RGBA data
 * returns: Image, finished, successful
 */
func (model *OverlappingModel) Iterate(iterations int) (image.Image, bool, bool) {
	finished := model.BaseModel.Iterate(model, iterations)
	return model.RenderImage(), finished, model.IsGenerationSuccessful()
}

/**
 * Retrieve the RGBA data
 * returns: Image, successful
 */
func (model *OverlappingModel) Generate() (image.Image, bool) {
	model.BaseModel.Generate(model)
	return model.RenderImage(), model.IsGenerationSuccessful()
}
