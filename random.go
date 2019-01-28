package wfc

func randomIndice(array []float64, r float64) int {
	sum := 0.0

	for _, n := range array {
		sum += n
	}

	if sum == 0.0 {
		for i := 0; i < len(array); i++ {
			array[i] = 1.0
		}

		sum = float64(len(array))
	}

	for i := 0; i < len(array); i++ {
		array[i] /= sum
	}

	i, x := 0, 0.0

	for i < len(array) {
		x += array[i]
		if r <= x {
			return i
		}
		i++
	}

	return 0
}
