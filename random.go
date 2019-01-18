package wavefunctioncollapse

import (
// "math/rand"
)

func randomIndice(array []int, r float64) int {
	sum := 0

	for n := range array {
		sum += n
	}

	if sum == 0 {
		for i := 0; i < len(array); i++ {
			array[i] = 1
		}

		sum = len(array)
	}

	for i := 0; i < len(array); i++ {
		array[i] /= sum
	}

	i, x := 0, 0

	for i < len(array) {
		x += array[i]
		if r <= float64(x) {
			return i
		}
		i++
	}

	return 0
}
