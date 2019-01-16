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
	n            int
	colors       []color.Color
	ground       int
	patterns     []Pattern
	propagator   [][][][]int
	fmxmn, fmymn int
}

/**
 * Pattern Type
 */
type Pattern []int

/**
 * Constructor
 */
func NewOverlappingModel(img image.Image, n, width, height int, periodicInput, periodicOutput bool, symmetry, ground int) *Model {

	// Initialize model
	model := &Model{}
	om := &OverlappingModel{}
	model.ConcreteModel = om
	om.n = n
	model.fmx = width
	model.fmy = height
	model.periodic = periodicOutput

	bounds := img.Bounds()
	dataWidth := bounds.Max.X
	dataHeight := bounds.Max.Y

	sample := make([][]int, dataWidth)
	for i := range sample {
		sample[i] = make([]int, dataHeight)
	}

	om.colors = make([]color.Color, 0)
	colorMap := make(map[color.Color]int)

	for y := 0; y < dataHeight; y++ {
		for x := 0; x < dataWidth; x++ {
			color := img.At(x, y)
			if _, ok := colorMap[color]; !ok {
				colorMap[color] = len(om.colors)
				om.colors = append(om.colors, color)
			}
			sample[x][y] = colorMap[color]
		}
	}

	// Extract patterns from input
	c := len(om.colors)
	w := int(math.Pow(float64(c), float64(n*n)))

	getPattern := func(f func(x, y int) int) Pattern {
		result := make(Pattern, n*n)
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				result[x+y*n] = f(x, y)
			}
		}
		return result
	}

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

	index := func(p Pattern) int {
		result := 0
		power := 1
		for i := 0; i < len(p); i++ {
			result += p[len(p)-1-i] * power
			power *= c
		}
		return result
	}

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
				ind := index(ps[k])
				if _, ok := weights[ind]; ok {
					weights[ind]++
				} else {
					weightsKeys = append(weightsKeys, ind)
					weights[ind] = 1
				}
			}
		}
	}

	model.t = len(weightsKeys)
	om.ground = ground

	om.patterns = make([]Pattern, model.t)
	model.stationary = make(Pattern, model.t)
	om.propagator = make([][][][]int, model.t)

	for wk, i := range weightsKeys {
		om.patterns[i] = patternFromIndex(wk)
		model.stationary[i] = weights[wk]
	}

	model.wave = make([][][]bool, model.fmx)
	model.changes = make([][]bool, model.fmx)

	for x := 0; x < model.fmx; x++ {
		model.wave[x] = make([][]bool, model.fmy)
		model.changes[x] = make([]bool, model.fmy)
		for y := 0; y < model.fmy; y++ {
			model.wave[x][y] = make([]bool, model.t)
			model.changes[x][y] = false
			for t := 0; t < model.t; t++ {
				model.wave[x][y][t] = true
			}
		}
	}

	// Propagate changes, checking for agreement
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

	for t := 0; t < model.t; t++ {
		om.propagator[t] = make([][][]int, 2*n-1)
		for x := 0; x < 2*n-1; x++ {
			om.propagator[t][x] = make([][]int, 2*n-1)
			for y := 0; y < 2*n-1; y++ {
				list := make([]int, 0)

				for t2 := 0; t2 < model.t; t2++ {
					if agrees(om.patterns[t], om.patterns[t2], x-n+1, y-n+1) {
						list = append(list, t2)
					}
				}

				om.propagator[t][x][y] = make([]int, len(list))

				for k := 0; k < len(list); k++ {
					om.propagator[t][x][y][k] = list[k]
				}
			}
		}
	}

	om.fmxmn = model.fmx - om.n
	om.fmymn = model.fmy - om.n

	return model
}

/**
 * OnBoundary
 */
func (model *OverlappingModel) OnBoundary(x, y int) bool {
	return !model.periodic && (x > model.fmxmn || y > model.fmymn)
}

/**
 * Propagate
 */
func (model *OverlappingModel) Propagate() bool {
	change := false
	startLoop := -model.n + 1
	endLoop := model.n

	for x := 0; x < model.fmx; x++ {
		for y := 0; y < model.fmy; y++ {
			if model.changes[x][y] {
				model.changes[x][y] = false
				for dx := startLoop; dx < endLoop; dx++ {
					for dy := startLoop; dy < endLoop; dy++ {
						sx := x + dx
						sy := y + dy

						if sx < 0 {
							sx += model.fmx
						} else if sx >= model.fmx {
							sx -= model.fmx
						}

						if sy < 0 {
							sy += model.fmy
						} else if sy >= model.fmy {
							sy -= model.fmy
						}

						if !model.periodic && (sx > model.fmxmn || sy > model.fmymn) {
							continue
						}

						allowed := &model.wave[sx][sy]

						for t := 0; t < model.t; t++ {
							if !allowed[t] {
								continue
							}

							b := false
							prop := model.propagator[t][model.n-1-dx][model.n-1-dy]
							for i := 0; i < len(prop) && !b; i++ {
								b = model.wave[x][y][prop[i]]
							}

							if !b {
								model.changes[sx][sy] = true
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
 * Clear the internal state
 */
func (model *OverlappingModel) Clear() {
	if model.ground != 0 {
		for x := 0; x < model.fmx; x++ {
			for t := 0; t < model.t; t++ {
				if t != ground {
					model.wave[x][model.fmy-1][t] = false
				}
			}

			model.changes[x][model.fmy-1] = true

			for y := 0; y < model.fmy-1; y++ {
				model.wave[x][y][model.ground] = false
				model.changes[x][y] = true
			}
		}

		for model.Propagate() {
			// Empty loop
		}
	}
}

/**
 * Retrieve the RGBA data
 */
func (model *OverlappingModel) Graphics() *image.Image {
	if model.isGenerationComplete() {
		return &NewCompleteImage(model)
	} else {
		return &NewIncompleteImage(model)
	}
}
