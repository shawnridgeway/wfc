package wavefunctioncollapse

import (
    "image"
    "math"
)

type Generator interface {
    Generate(rand RandNumGen) (image.Image, bool)
}

type AppliedAlgorithm interface {
    Generator
    OnBoundary(x, y int) bool
    Propagate() bool
    Clear()
}

type BaseModel struct {
    InitiliazedField, GenerationComplete bool       // Flags for status
    Wave                                 [][][]bool // All possible patterns (t) that could fit coordinates (x, y)
    Changes                              [][]bool   // Changes made in interation of propagation
    Stationary                           []int      // Array of weights (by frequency) for each pattern (matches index in patterns field)
    T                                    int        // Count of patterns
    Periodic                             bool       // Output is periodic (ie tessellates)
    Fmx, Fmy                             int        // Width and height of output
    Rng                                  RandNumGen // Random number generator supplied at generation time
}

type RandNumGen func() float64

/**
 * Observe
 * returns: completed (bool), success (bool)
 */
func (baseModel *BaseModel) Observe(specificModel AppliedAlgorithm) (bool, bool) {
    min := 1000.0
    argminx := -1
    argminy := -1
    distribution := make([]int, baseModel.T)

    for x := 0; x < baseModel.Fmx; x++ {
        wavex := baseModel.Wave[x]
        for y := 0; y < baseModel.Fmy; y++ {
            if specificModel.OnBoundary(x, y) {
                continue
            }

            sum := 0

            for t := 0; t < baseModel.T; t++ {
                if wavex[y][t] {
                    distribution[t] = baseModel.Stationary[t]
                } else {
                    distribution[t] = 0
                }
                sum += distribution[t]
            }

            if sum == 0 {
                return true, false
            }

            for t := 0; t < baseModel.T; t++ {
                distribution[t] /= sum
            }

            entropy := 0.0

            for i := 0; i < len(distribution); i++ {
                if distribution[i] > 0 {
                    entropy += -float64(distribution[i]) * math.Log(float64(distribution[i]))
                }
            }

            noise := 0.000001 * baseModel.Rng()

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

    for t := 0; t < baseModel.T; t++ {
        if baseModel.Wave[argminx][argminy][t] {
            distribution[t] = baseModel.Stationary[t]
        } else {
            distribution[t] = 0
        }
    }

    r := randomIndice(distribution, baseModel.Rng())

    for t := 0; t < baseModel.T; t++ {
        baseModel.Wave[argminx][argminy][t] = (t == r)
    }

    baseModel.Changes[argminx][argminy] = true

    return false, false
}

/**
 * Execute a single iteration
 * returns: completed (bool), success (bool)
 */
func (baseModel *BaseModel) SingleIteration(specificModel AppliedAlgorithm) (bool, bool) {
    completed, success := baseModel.Observe(specificModel)

    if completed {
        baseModel.GenerationComplete = success
        return true, success
    }

    for specificModel.Propagate() {
        // Empty loop
    }

    return false, false
}

/**
 * Execute a fixed number of iterations. Stop when the generation is successful or reaches a contradiction.
 */
func (baseModel *BaseModel) Iterate(specificModel AppliedAlgorithm, iterations int) bool {
    if !baseModel.InitiliazedField {
        specificModel.Clear()
    }

    for i := 0; i < iterations || iterations == 0; i++ {
        completed, success := baseModel.SingleIteration(specificModel)
        if completed {
            return success
        }
    }

    return true
}

/**
 * Execute a complete new generation
 */
func (baseModel *BaseModel) Generate(specificModel AppliedAlgorithm) bool {
    specificModel.Clear()
    for {
        completed, success := baseModel.SingleIteration(specificModel)
        if completed {
            return success
        }
    }
}

/**
 * Check whether the previous generation completed successfully
 */
func (baseModel *BaseModel) IsGenerationComplete(specificModel AppliedAlgorithm) bool {
    return baseModel.GenerationComplete
}

/**
 * Clear the internal state to start a new generation
 */
func (baseModel *BaseModel) ClearBase(specificModel AppliedAlgorithm) {
    for x := 0; x < baseModel.Fmx; x++ {
        for y := 0; y < baseModel.Fmy; y++ {
            for t := 0; t < baseModel.T; t++ {
                baseModel.Wave[x][y][t] = true
            }
            baseModel.Changes[x][y] = false
        }
    }
    baseModel.InitiliazedField = true
    baseModel.GenerationComplete = false
}
