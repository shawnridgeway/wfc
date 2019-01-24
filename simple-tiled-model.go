package wavefunctioncollapse

import (
	// "fmt"
	"image"
	"image/color"
)

/**
 * SimpleTiledModel Type
 */
type SimpleTiledModel struct {
	*BaseModel               // Underlying model of generic Wave Function Collapse algorithm
	TileSize   int           // The size in pixels of the length and height of each tile
	Tiles      []TilePattern // List of all possible tiles as images, including inversions
	Propagator [][][]bool    // All possible connections between tiles
}

// Parsed data supplied by user
type SimpleTiledData struct {
	Unique    bool       // False if each tile can have variants. (Default to false?)
	TileSize  int        // Default to 16
	Tiles     []Tile     // List of all possible tiles, not including inversions
	Neighbors []Neighbor // List of possible connections between tiles
}

// Raw information on a tile
type Tile struct {
	Name     string        // Name used to identify the tile
	Symmetry string        // Default to ""
	Weight   float64       // Default to 1
	Variants []image.Image // Preloaded image for the tile
}

// Information on which tiles can be neighbors
type Neighbor struct {
	Left     string // Mathces Tile.Name
	LeftNum  int    // Default to 0
	Right    string // Mathces Tile.Name
	RightNum int    // Default to 0
}

// Flat array of colors in a tile
type TilePattern []color.Color

// Tile inversion function type
type Inversion func(int) int

/**
 * NewSimpleTiledModel
 * @param {object} data Tiles and constraints definitions
 * @param {int} width The width of the generation, in terms of tiles (not pixels)
 * @param {int} height The height of the generation, in terms of tiles (not pixels)
 * @param {bool} periodic Whether the source image is to be considered as periodic / as a repeatable texture
 * @return *SimpleTiledModel A pointer to a new copy of the model
 */
func NewSimpleTiledModel(data SimpleTiledData, width, height int, periodic bool) *SimpleTiledModel {

	// Initialize model
	model := &SimpleTiledModel{BaseModel: &BaseModel{}}
	model.Fmx = width
	model.Fmy = height
	model.Periodic = periodic
	model.TileSize = data.TileSize
	model.Tiles = make([]TilePattern, 0)
	model.Stationary = make([]float64, 0)

	firstOccurrence := make(map[string]int)
	action := make([][]int, 0)

	tile := func(transformer func(x, y int) color.Color) TilePattern {
		result := make(TilePattern, model.TileSize*model.TileSize)
		for y := 0; y < model.TileSize; y++ {
			for x := 0; x < model.TileSize; x++ {
				result[x+y*model.TileSize] = transformer(x, y)
			}
		}
		return result
	}

	rotate := func(p TilePattern) TilePattern {
		return tile(func(x, y int) color.Color {
			return p[model.TileSize-1-y+x*model.TileSize]
		})
	}

	for i := 0; i < len(data.Tiles); i++ {
		currentTile := data.Tiles[i]
		var cardinality int
		var inversion1, inversion2 Inversion

		switch currentTile.Symmetry {
		case "L":
			cardinality = 4
			inversion1 = func(i int) int {
				return (i + 1) % 4
			}
			inversion2 = func(i int) int {
				if i%2 == 0 {
					return i + 1
				}
				return i - 1
			}
		case "T":
			cardinality = 4
			inversion1 = func(i int) int {
				return (i + 1) % 4
			}
			inversion2 = func(i int) int {
				if i%2 == 0 {
					return i
				}
				return 4 - i
			}
		case "I":
			cardinality = 2
			inversion1 = func(i int) int {
				return 1 - i
			}
			inversion2 = func(i int) int {
				return i
			}
		case "\\":
			cardinality = 2
			inversion1 = func(i int) int {
				return 1 - i
			}
			inversion2 = func(i int) int {
				return 1 - i
			}
		case "X":
			cardinality = 1
			inversion1 = func(i int) int {
				return i
			}
			inversion2 = func(i int) int {
				return i
			}
		default:
			cardinality = 1
			inversion1 = func(i int) int {
				return i
			}
			inversion2 = func(i int) int {
				return i
			}
		}

		model.T = len(action)
		firstOccurrence[currentTile.Name] = model.T

		for t := 0; t < cardinality; t++ {
			action = append(action, []int{
				model.T + t,
				model.T + inversion1(t),
				model.T + inversion1(inversion1(t)),
				model.T + inversion1(inversion1(inversion1(t))),
				model.T + inversion2(t),
				model.T + inversion2(inversion1(t)),
				model.T + inversion2(inversion1(inversion1(t))),
				model.T + inversion2(inversion1(inversion1(inversion1(t)))),
			})
		}

		if data.Unique {
			for t := 0; t < cardinality; t++ {
				img := currentTile.Variants[t]
				model.Tiles = append(model.Tiles, tile(func(x, y int) color.Color {
					return img.At(x, y)
				}))
			}
		} else {
			img := currentTile.Variants[0]
			model.Tiles = append(model.Tiles, tile(func(x, y int) color.Color {
				return img.At(x, y)
			}))

			for t := 1; t < cardinality; t++ {
				model.Tiles = append(model.Tiles, rotate(model.Tiles[model.T+t-1]))
			}
		}

		for t := 0; t < cardinality; t++ {
			model.Stationary = append(model.Stationary, currentTile.Weight)
		}
	}

	model.T = len(action)
	model.Propagator = make([][][]bool, 4)

	for i := 0; i < 4; i++ {
		model.Propagator[i] = make([][]bool, model.T)
		for t := 0; t < model.T; t++ {
			model.Propagator[i][t] = make([]bool, model.T)
			for t2 := 0; t2 < model.T; t2++ {
				model.Propagator[i][t][t2] = false
			}
		}
	}

	model.Wave = make([][][]bool, model.Fmx)
	model.Changes = make([][]bool, model.Fmx)
	for x := 0; x < model.Fmx; x++ {
		model.Wave[x] = make([][]bool, model.Fmy)
		model.Changes[x] = make([]bool, model.Fmy)
		for y := 0; y < model.Fmy; y++ {
			model.Wave[x][y] = make([]bool, model.T)
		}
	}

	for i := 0; i < len(data.Neighbors); i++ {
		neighbor := data.Neighbors[i]

		l := action[firstOccurrence[neighbor.Left]][neighbor.LeftNum]
		d := action[l][1]
		r := action[firstOccurrence[neighbor.Right]][neighbor.RightNum]
		u := action[r][1]

		model.Propagator[0][r][l] = true
		model.Propagator[0][action[r][6]][action[l][6]] = true
		model.Propagator[0][action[l][4]][action[r][4]] = true
		model.Propagator[0][action[l][2]][action[r][2]] = true

		model.Propagator[1][u][d] = true
		model.Propagator[1][action[d][6]][action[u][6]] = true
		model.Propagator[1][action[u][4]][action[d][4]] = true
		model.Propagator[1][action[d][2]][action[u][2]] = true
	}

	for t := 0; t < model.T; t++ {
		for t2 := 0; t2 < model.T; t2++ {
			model.Propagator[2][t][t2] = model.Propagator[0][t2][t]
			model.Propagator[3][t][t2] = model.Propagator[1][t2][t]
		}
	}

	return model
}

/**
 * OnBoundary
 */
func (model *SimpleTiledModel) OnBoundary(x, y int) bool {
	return false
}

/**
 * Propagate
 * return: bool, change occured in this iteration
 */
func (model *SimpleTiledModel) Propagate() bool {
	change := false

	for x2 := 0; x2 < model.Fmx; x2++ {
		for y2 := 0; y2 < model.Fmy; y2++ {
			for d := 0; d < 4; d++ {
				x1 := x2
				y1 := y2

				if d == 0 {
					if x2 == 0 {
						if !model.Periodic {
							continue
						} else {
							x1 = model.Fmx - 1
						}
					} else {
						x1 = x2 - 1
					}
				} else if d == 1 {
					if y2 == model.Fmy-1 {
						if !model.Periodic {
							continue
						} else {
							y1 = 0
						}
					} else {
						y1 = y2 + 1
					}
				} else if d == 2 {
					if x2 == model.Fmx-1 {
						if !model.Periodic {
							continue
						} else {
							x1 = 0
						}
					} else {
						x1 = x2 + 1
					}
				} else {
					if y2 == 0 {
						if !model.Periodic {
							continue
						} else {
							y1 = model.Fmy - 1
						}
					} else {
						y1 = y2 - 1
					}
				}

				if !model.Changes[x1][y1] {
					continue
				}

				for t2 := 0; t2 < model.T; t2++ {
					if model.Wave[x2][y2][t2] {
						b := false

						for t1 := 0; t1 < model.T && !b; t1++ {
							if model.Wave[x1][y1][t1] {
								b = model.Propagator[d][t2][t1]
							}
						}

						if !b {
							model.Wave[x2][y2][t2] = false
							model.Changes[x2][y2] = true
							change = true
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
func (model *SimpleTiledModel) Clear() {
	model.ClearBase(model)
}

/**
 * Create a GeneratedImage holding the data for a complete image
 */
func (model *SimpleTiledModel) RenderCompleteImage() GeneratedImage {
	output := make([][]color.Color, model.Fmx)
	for i := range output {
		output[i] = make([]color.Color, model.Fmy)
	}
	for y := 0; y < model.Fmy; y++ {
		for x := 0; x < model.Fmx; x++ {
			for yt := 0; yt < model.TileSize; yt++ {
				for xt := 0; xt < model.TileSize; xt++ {
					for t := 0; t < model.T; t++ {
						if model.Wave[x][y][t] {
							output[x*model.TileSize+xt][y*model.TileSize+yt] = model.Tiles[t][yt*model.TileSize+xt]
							break
						}
					}
				}
			}
		}
	}
	return GeneratedImage{output}
}

/**
 * Create a GeneratedImage holding the data for an incomplete image
 */
func (model *SimpleTiledModel) RenderIncompleteImage() GeneratedImage {
	output := make([][]color.Color, model.Fmx)
	for i := range output {
		output[i] = make([]color.Color, model.Fmy)
	}
	for y := 0; y < model.Fmy; y++ {
		for x := 0; x < model.Fmx; x++ {
			amount := 0
			sum := 0.0
			for t := 0; t < len(model.Wave[x][y]); t++ {
				if model.Wave[x][y][t] {
					amount += 1
					sum += model.Stationary[t]
				}
			}
			for yt := 0; yt < model.TileSize; yt++ {
				for xt := 0; xt < model.TileSize; xt++ {
					if amount == model.T {
						output[x*model.TileSize+xt][y*model.TileSize+yt] = color.RGBA{127, 127, 127, 255}
					} else {
						sR, sG, sB, sA := 0.0, 0.0, 0.0, 0.0
						for t := 0; t < model.T; t++ {
							if model.Wave[x][y][t] {
								r, g, b, a := model.Tiles[t][yt*model.TileSize+xt].RGBA()
								sR += float64(r) * model.Stationary[t]
								sG += float64(g) * model.Stationary[t]
								sB += float64(b) * model.Stationary[t]
								sA += float64(a) * model.Stationary[t]
							}
						}
						uR := uint8(int(sR/sum) >> 8)
						uG := uint8(int(sG/sum) >> 8)
						uB := uint8(int(sB/sum) >> 8)
						uA := uint8(int(sA/sum) >> 8)
						output[x*model.TileSize+xt][y*model.TileSize+yt] = color.RGBA{uR, uG, uB, uA}
					}
				}
			}
		}
	}
	return GeneratedImage{output}
}

/**
 * Retrieve the RGBA data
 * returns: Image
 */
func (model *SimpleTiledModel) RenderImage() GeneratedImage {
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
func (model *SimpleTiledModel) Iterate(iterations int) (image.Image, bool, bool) {
	finished := model.BaseModel.Iterate(model, iterations)
	return model.RenderImage(), finished, model.IsGenerationSuccessful()
}

/**
 * Retrieve the RGBA data
 * returns: Image, successful
 */
func (model *SimpleTiledModel) Generate() (image.Image, bool) {
	model.BaseModel.Generate(model)
	return model.RenderImage(), model.IsGenerationSuccessful()
}
