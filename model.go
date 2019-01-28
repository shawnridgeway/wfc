package wfc

import (
    // "fmt"
    "image"
    "math"
    "math/rand"
    "time"
)

type Iterator interface {
    Iterate(iterations int) (image.Image, bool, bool)
}

type Generator interface {
    Generate() (image.Image, bool)
}

type AppliedAlgorithm interface {
    Iterator
    Generator
    OnBoundary(x, y int) bool
    Propagate() bool
    Clear()
}

type BaseModel struct {
    InitiliazedField     bool           // Generation Initialized
    RngSet               bool           // Random number generator set by user
    GenerationSuccessful bool           // Generation has run into a contradiction
    Wave                 [][][]bool     // All possible patterns (t) that could fit coordinates (x, y)
    Changes              [][]bool       // Changes made in interation of propagation
    Stationary           []float64      // Array of weights (by frequency) for each pattern (matches index in patterns field)
    T                    int            // Count of patterns
    Periodic             bool           // Output is periodic (ie tessellates)
    Fmx, Fmy             int            // Width and height of output
    Rng                  func() float64 // Random number generator supplied at generation time
}

/**
 * Observe
 * returns: finished (bool)
 */
func (baseModel *BaseModel) Observe(specificModel AppliedAlgorithm) bool {
    min := 1000.0
    argminx := -1
    argminy := -1
    distribution := make([]float64, baseModel.T)

    // Find the point with minimum entropy (adding a little noise for randomness)
    for x := 0; x < baseModel.Fmx; x++ {
        for y := 0; y < baseModel.Fmy; y++ {
            if specificModel.OnBoundary(x, y) {
                continue
            }

            sum := 0.0

            for t := 0; t < baseModel.T; t++ {
                if baseModel.Wave[x][y][t] {
                    distribution[t] = baseModel.Stationary[t]
                } else {
                    distribution[t] = 0.0
                }
                sum += distribution[t]
            }

            if sum == 0.0 {
                baseModel.GenerationSuccessful = false
                return true // finished, unsuccessful
            }

            for t := 0; t < baseModel.T; t++ {
                distribution[t] /= sum
            }

            entropy := 0.0

            for i := 0; i < len(distribution); i++ {
                if distribution[i] > 0.0 {
                    entropy += -distribution[i] * math.Log(distribution[i])
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
        baseModel.GenerationSuccessful = true
        return true // finished, successful
    }

    for t := 0; t < baseModel.T; t++ {
        if baseModel.Wave[argminx][argminy][t] {
            distribution[t] = baseModel.Stationary[t]
        } else {
            distribution[t] = 0.0
        }
    }

    r := randomIndice(distribution, baseModel.Rng())

    for t := 0; t < baseModel.T; t++ {
        baseModel.Wave[argminx][argminy][t] = (t == r)
    }

    baseModel.Changes[argminx][argminy] = true

    return false // Not finished yet
}

/**
 * Execute a single iteration
 * returns: finished (bool)
 */
func (baseModel *BaseModel) SingleIteration(specificModel AppliedAlgorithm) bool {
    finished := baseModel.Observe(specificModel)

    if finished {
        return true
    }

    for specificModel.Propagate() {
        // Empty loop
    }

    return false // Not finished yet
}

/**
 * Execute a fixed number of iterations. Stop when the generation succeedes or fails.
 */
func (baseModel *BaseModel) Iterate(specificModel AppliedAlgorithm, iterations int) bool {
    if !baseModel.InitiliazedField {
        specificModel.Clear()
    }

    for i := 0; i < iterations; i++ {
        finished := baseModel.SingleIteration(specificModel)
        if finished {
            return true
        }
    }
    return false // Not finished yet
}

/**
 * Execute a complete new generation until success or failure.
 */
func (baseModel *BaseModel) Generate(specificModel AppliedAlgorithm) {
    specificModel.Clear()
    for {
        finished := baseModel.SingleIteration(specificModel)
        if finished {
            return
        }
    }
}

/**
 * Check whether the generation completed successfully
 */
func (baseModel *BaseModel) IsGenerationSuccessful() bool {
    return baseModel.GenerationSuccessful
}

/**
 * Set the seed for the random number generator. Useful for a stable testing environment.
 */
func (baseModel *BaseModel) SetSeed(seed int64) {
    baseModel.Rng = rand.New(rand.NewSource(seed)).Float64
    baseModel.RngSet = true
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
    if !baseModel.RngSet {
        baseModel.Rng = rand.New(rand.NewSource(time.Now().UnixNano())).Float64
    }
    baseModel.InitiliazedField = true
    baseModel.GenerationSuccessful = false
}
