package wavefunctioncollapse

import (
	// "fmt"
	"image/color"
	"math"
)

type OverlappingModel struct {
	*Model
	n int
	// periodic bool
	// fmx, fmy int
	colors color.RGBA
	// t int
	ground int
	patterns Pattern
	// stationary Pattern
	propagator [][][][]int
	// wave [][][]bool
	// changes [][]bool
	fmxmn, fmymn int
}

type Pattern []int

type OutputFormat []uint32

/**
 * Constructor
 */
func NewOverlappingModel(data OutputFormat, dataWidth, dataHeight, n, width, height int, periodicInput, periodicOutput bool, symmetry, ground int) *OverlappingModel {

	// Initialize model
	model = OverlappingModel{}
	model.n = n
	model.fmx = width
    model.fmy = height
    model.periodic = periodicOutput

    sample := make([][]int, dataWidth)
    for i := range sample {
    	sample[i] = make([]int, dataHeight)
    }

    model.colors := make([]color.RGBA, 0)
    colorMap := make(map[color.RGBA]int)

    for y := 0; y < dataHeight; y++ {
    	for x := 0; x < dataWidth; x++ {
    		indexPixel := (y * dataHeight + x) * 4
    		color := color.RGBA{data[indexPixel], data[indexPixel + 1], data[indexPixel + 2], data[indexPixel + 3]}
    		if _, ok := colorMap; ok {
    			colorMap[color] = len(model.colors)
    			model.colors = append(model.colors, color)
    		}
    		sample[x][y] = colorMap[color]
    	}
    }

    // Extract patterns from input
    c := len(model.colors)
    w := math.Pow(c, n*n)

    getPattern := func(f func(x, y int) int) Pattern {
    	result := make(Pattern, n*n)
    	for y := 0; y < n; y++ {
    		for x := 0; x < n; x++ {
    			result[x + y * n] = f(x, y)
    		}
    	}
    	return result
    }

    patternFromSample := func(x, y int) Pattern {
    	return getPattern(func(dx, dy int) int {
    		return sample[(x + dx) % dataWidth][(y + dy) % dataHeight]
		})
    }

    rotate := func(p Pattern) int {
    	return getPattern(func(x, y int) int {
    		return p[n - 1 - y + x * n]
		})
    }

    reflect := func(p Pattern) int {
    	return getPattern(func(x, y int) int {
    		return p[n - 1 - x + y * n]
		})
    }

    index := func(p Pattern) int {
    	result := 0
    	power := 1
    	for i := 0; i < len(p); i++ {
    		result += p[len(p) - 1 - i] * power
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
    		ps := make([]int, 8, 8)
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
            		weightsKeys = append(weightsKeys, ind);
                    weights[ind] = 1
            	}
            }
    	}
    }

    model.t = len(weightsKeys)
    model.ground = ground

    model.patterns = make(Pattern, model.t)
    model.stationary = make(Pattern, model.t)
    model.propagator = make([][][][]int, model.t)

    for wk, i := range weightsKeys {
    	model.patterns[i] = patternFromIndex(wk)
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
    agrees := func(p1, p2 Pattern, dx, dy int) {
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
    			if p1[x + n * y] != p2[x - dx + n * (y - dy)] {
    				return false
    			}
    		}
    	}

    	return true
    }

    for t := 0; t < model.t; t++ {
    	model.propagator[t] = make([][][]int, 2 * n - 1)
    	for x := 0; x < 2 * n - 1; x++ {
    		model.propagator[t][x] = make([][]int, 2 * n - 1)
    		for y := 0; y < 2 * n - 1; y++ {
    			list := make([]int, 0)

    			for t2 := 0; t2 < model.t; t2++ {
    				if agrees(model.patterns[t], model.patterns[t2], x - n + 1, y - n + 1) {
    					list = append(list, t2)
    				}
    			}

    			model.propagator[t][x][y] = make([]int, len(list))

    			for k := 0; k < len(list); k++ {
    				model.propagator[t][x][y][k] = list[k]
    			}
    		}
    	}
    }

    model.fmxmn = model.fmx - model.n
    model.fmymn = model.fmy - model.n

    return &model
}

/**
 * Shorthand Constructors
 */
func NewOverlappingModel(data OutputFormat, dataWidth, dataHeight, n, width, height int) *OverlappingModel {
	return NewOverlappingModel(data, dataWidth, dataHeight, n, width, height, false, false, 1, 0)
}

func NewOverlappingModel(data OutputFormat, dataWidth, dataHeight, n int) *OverlappingModel {
	return NewOverlappingModel(data, dataWidth, dataHeight, n, dataWidth, dataHeight)
}

func NewOverlappingModel(data OutputFormat, dataWidth, dataHeight int) *OverlappingModel {
	return NewOverlappingModel(data, dataWidth, dataHeight, 0)
}

/**
 */
func (model *OverlappingModel) onBoundary(x, y int) bool {
	return !model.periodic && (x > model.fmxmn || y > model.fmymn)
}

/**
 */
func (model *OverlappingModel) propagate() bool {
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
                        	prop := model.propagator[t][model.n - 1 - dx][model.n - 1 - dy]
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
	model.Model.Clear()
	if model.ground != 0 {
		for x := 0; x < model.fmx; x++ {
			for t := 0; t < model.t; t++ {
				if t != ground {
					model.wave[x][model.fmy - 1][t] = false
				}
			}

			model.changes[x][model.fmy - 1] = true

			for y := 0; y < model.fmy - 1; y++ {
				model.wave[x][y][model.ground] = false
                model.changes[x][y] = true
			}
		}

		for model.propagate() {
			// Empty loop
		}
	}
}

/**
 * Set the RGBA data for a complete generation in a given array
 */
func (model *OverlappingModel) graphicsComplete() OutputFormat {
	array := make(OutputFormat, model.fmx * model.fmy * 4)

	for y := 0; y < model.fmy; y++ {
		for x := 0; x < model.fmx; x++ {
			pixelIndex := (y * model.fmx + x) * 4
			for t := 0; t < model.t; t++ {
				if model.wave[x][y][t] {
					color := model.colors[model.patterns[t][0]]
					array[pixelIndex] = color.R
                    array[pixelIndex + 1] = color.G
                    array[pixelIndex + 2] = color.B
                    array[pixelIndex + 3] = color.A
                    break
				}
			}
		}
	}

	return array
}

/**
 * Set the RGBA data for an incomplete generation in a given array
 */
func (model *OverlappingModel) graphicsIncomplete() OutputFormat {
	array := make(OutputFormat, model.fmx * model.fmy * 4)

	for y := 0; y < model.fmy; y++ {
		for x := 0; x < model.fmx; x++ {
			contributorNumber, r, g, b, a := 0, 0, 0, 0, 0
			pixelIndex := (y * model.fmx + x) * 4
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

                    if !model.periodic && (sx + model.n > model.fmx || sy + model.n > model.fmy) {
                        continue
                    }

                    for t := 0; t < model.t; t++ {
                    	if model.wave[sx][sy][t] {
                    		contributorNumber++
                    		color = model.colors[model.patterns[t][dx + dy * model.n]]
                    		r += color.R
                            g += color.G
                            b += color.B
                            a += color.A
                    	}
                    }
				}
			}

			array[pixelIndex] = r / contributorNumber
            array[pixelIndex + 1] = g / contributorNumber
            array[pixelIndex + 2] = b / contributorNumber
            array[pixelIndex + 3] = a / contributorNumber
		}
	}

	return array
}

/**
 * Retrieve the RGBA data
 */
func (model *OverlappingModel) graphics() OutputFormat {
	if model.isGenerationComplete() {
		return model.graphicsComplete()
	} else {
		return model.graphicsIncomplete()
	}
}


