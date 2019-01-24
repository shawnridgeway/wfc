package wavefunctioncollapse

import (
	"github.com/shawnridgeway/wavefunctioncollapse/internal/testutils"
	"testing"
)

func TestNewOverlappingModel(t *testing.T) {
	// Set test parameters
	filename := "flowers.png"
	periodicInput := true
	periodicOutput := true
	hasGround := true
	width := 48
	height := 48
	n := 3
	symetry := 2
	seed := int64(42)

	// Load input sample image
	inputImg, err := testutils.LoadImage("internal/input/" + filename)
	if err != nil {
		panic(err)
	}

	// Generate output image
	model := NewOverlappingModel(inputImg, n, width, height, periodicInput, periodicOutput, symetry, hasGround)
	model.SetSeed(seed)
	outputImg, success := model.Generate()
	if !success {
		t.Log("Failed to generate image on the first try.")
		t.FailNow()
	}

	// Save output
	// err = testutils.SaveImage("internal/target/"+filename, outputImg)
	// if err != nil {
	// 	panic(err)
	// }

	// Test that files match
	targetImg, err := testutils.LoadImage("internal/target/" + filename)
	if err != nil {
		panic(err)
	}
	areEqual := testutils.CompareImages(outputImg, targetImg)
	if !areEqual {
		t.Log("Output image is not the same as the target image.")
		t.FailNow()
	}
}
