package wavefunctioncollapse

import (
	"github.com/shawnridgeway/wavefunctioncollapse/internal/testutils"
	"image"
	"testing"
)

func overlappingTest(t *testing.T, filename, snapshotFilename string, iterations int) {
	// Set test parameters
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
	var outputImg image.Image
	success, finished := false, false
	model := NewOverlappingModel(inputImg, n, width, height, periodicInput, periodicOutput, symetry, hasGround)
	model.SetSeed(seed)
	if iterations == -1 {
		outputImg, success = model.Generate()
		if !success {
			t.Log("Failed to generate image on the first try.")
			t.FailNow()
		}
	} else {
		outputImg, finished, _ = model.Iterate(iterations)
		if finished {
			t.Log("Test for incomplete state actually finished.")
			t.FailNow()
		}
	}

	// Save output
	// err = testutils.SaveImage("internal/snapshots/"+snapshotFilename, outputImg)
	// if err != nil {
	// 	panic(err)
	// }

	// Test that files match
	snapshotImg, err := testutils.LoadImage("internal/snapshots/" + snapshotFilename)
	if err != nil {
		panic(err)
	}
	areEqual := testutils.CompareImages(outputImg, snapshotImg)
	if !areEqual {
		t.Log("Output image is not the same as the snapshot image.")
		t.FailNow()
	}
}

func TestOverlappingGenerationCompletes(t *testing.T) {
	overlappingTest(t, "flowers.png", "flowers.png", -1)
}

func TestOverlappingIterationIncomplete(t *testing.T) {
	overlappingTest(t, "flowers.png", "flowers_incomplete.png", 5)
}
