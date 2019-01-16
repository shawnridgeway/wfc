package wavefunctioncollapse

import (
	"image"
	"math"
)

type ConcreteModel interface {
	OnBoundary(x, y int) bool
	Propagate() bool
	Clear()
	Graphics() *image.Image
}

type Model struct {
	ConcreteModel
	initiliazedField, generationComplete bool
	wave                                 [][][]bool
	changes                              [][]bool
	stationary                           Pattern
	t                                    int
	periodic                             bool
	fmx, fmy                             int
}

type RandNumGen func() float64

/**
 * observe
 * returns: completed (bool), success (bool)
 */
func (model *Model) observe(rng RandNumGen) (bool, bool) {
	min := 1000.0
	argminx := -1
	argminy := -1
	distribution := make([]int, model.t)

	for x := 0; x < model.fmx; x++ {
		wavex := model.wave[x]
		for y := 0; y < model.fmy; y++ {
			if model.OnBoundary(x, y) {
				continue
			}

			sum := 0

			for t := 0; t < model.t; t++ {
				if wavex[y][t] {
					distribution[t] = model.stationary[t]
				} else {
					distribution[t] = 0
				}
				sum += distribution[t]
			}

			if sum == 0 {
				return true, false
			}

			for t := 0; t < model.t; t++ {
				distribution[t] /= sum
			}

			entropy := 0.0

			for i := 0; i < len(distribution); i++ {
				if distribution[i] > 0 {
					entropy += -float64(distribution[i]) * math.Log(float64(distribution[i]))
				}
			}

			noise := 0.000001 * rng()

			if entropy > 0 && entropy+noise < min {
				min = entropy + noise
				argminx = x
				argminy = y
			}
		}
	}

	if argminx == -1 && argminy == -1 {
		return true, true
	}

	for t := 0; t < model.t; t++ {
		if model.wave[argminx][argminy][t] {
			distribution[t] = model.stationary[t]
		} else {
			distribution[t] = 0
		}
	}

	r := randomIndice(distribution, rng())

	for t := 0; t < model.t; t++ {
		model.wave[argminx][argminy][t] = (t == r)
	}

	model.changes[argminx][argminy] = true

	return false, false
}

/**
 * Execute a single iteration
 * returns: completed (bool), success (bool)
 */
func (model *Model) singleIteration(rng RandNumGen) (bool, bool) {
	completed, success := model.observe(rng)

	if completed {
		model.generationComplete = success
		return true, success
	}

	for model.Propagate() {
		// Empty loop
	}

	return false, false
}

/**
 * Execute a fixed number of iterations. Stop when the generation is successful or reaches a contradiction.
 */
func (model *Model) Iterate(iterations int, rng RandNumGen) bool {
	if !model.initiliazedField {
		model.Clear()
	}

	for i := 0; i < iterations || iterations == 0; i++ {
		completed, success := model.singleIteration(rng)
		if completed {
			return success
		}
	}

	return true
}

/**
 * Execute a complete new generation
 */
func (model *Model) Generate(rng RandNumGen) bool {
	model.Clear()
	for {
		completed, success := model.singleIteration(rng)
		if completed {
			return success
		}
	}
}

/**
 * Check whether the previous generation completed successfully
 */
func (model *Model) IsGenerationComplete() bool {
	return model.generationComplete
}

/**
 * Clear the internal state to start a new generation
 */
func (model *Model) Clear() {
	for x := 0; x < model.fmx; x++ {
		for y := 0; y < model.fmy; y++ {
			for t := 0; t < model.t; t++ {
				model.wave[x][y][t] = true
			}
			model.changes[x][y] = false
		}
	}
	model.initiliazedField = true
	model.generationComplete = false
	model.ConcreteModel.Clear()
}
