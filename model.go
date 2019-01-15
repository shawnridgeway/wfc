package wavefunctioncollapse

import (
	"math"
)

type Model struct {
	initiliazedField, generationComplete bool
	wave [][][]bool
	changes [][]bool
	stationary Pattern
	t int
	periodic bool
	fmx, fmy int
}

type RandNumGen func(array []int, r int) int

type ObservationResult bool | nil

/**
 * 
 */
func (model *Model) observe(rng RandNumGen) ObservationResult {
	min := 1000
    argminx := -1
    argminy := -1
    distribution := make([]int, model.t)

    for x := 0; x < model.fmx; x++ {
    	wavex := &model.wave[x]
    	for y := 0; y < model.fmy; y++ {
    		if model.onBoundary(x, y) {
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
    			return false
    		}

    		for t := 0; t < model.t; t++ {
    			distribution[t] /= sum
    		}

    		entropy := 0

    		for i := 0; i < len(distribution); i++ {
    			if distribution[i] > 0 {
    				entropy += -distribution[i] * math.Log(distribution[i])
    			}
    		}

    		noise := 0.000001 * rng()

    		if entropy > 0 && entropy + noise < min {
    			min = entropy + noise
    			argminx = x
    			argminy = y
    		}
    	}
    }

    if argminx == -1 && argminy == -1 {
    	return true
    }

    for t := 0; t < model.t; t++ {
    	if model.wave[argminx][argminy][t] {
    		distribution[t] =  model.stationary[t]
    	} else {
    		distribution[t] = 0
    	}
    }

    r := randomIndice(distribution, rng())

    for t := 0; t < model.t; t++ {
    	model.wave[argminx][argminy][t] = (t == r)
    }

    model.changes[argminx][argminy] = true

    return nil
}

/**
 * Execute a single iteration
 */
func (model *Model) singleIteration(rng RandNumGen) ObservationResult {
	result := model.observe(rng)

	if rng != nil {
		model.generationComplete = result
		return result
	}

	for model.propagate() {
		// Empty loop
	}

    return nil
}

/**
 * Execute a fixed number of iterations. Stop when the generation is successful or reaches a contradiction.
 */
func (model *Model) Iterate(iterations int, rng RandNumGen) bool {
	if !model.initiliazedField {
		model.Clear()
	}

 	for i := 0; i < iterations || iterations == 0; i++ {
 		result := model.singleIteration(rng)
 		if result != nil {
 			return result
 		}
 	}

 	return true
}

func (model *Model) Iterate(iterations int) bool {
	rand.Seed(time.Now().UnixNano())
	rng := rand.Float64
	return model.Generate(iterations, rng)
}

func (model *Model) Iterate() bool {
	return model.Generate(0)
}

/**
 * Execute a complete new generation
 */
func (model *Model) Generate(rng RandNumGen) bool {
	model.Clear()
	for {
		result := model.singleIteration(rng)
		if result != nil {
			return result
		}
	}
}

func (model *Model) Generate() bool {
	rand.Seed(time.Now().UnixNano())
	rng := rand.Float64
	return model.Generate(rng)
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
}

