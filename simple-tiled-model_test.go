package wavefunctioncollapse

import (
	"encoding/json"
	"github.com/shawnridgeway/wavefunctioncollapse/internal/testutils"
	"image"
	"io/ioutil"
	"strconv"
	"testing"
)

// Parsed data supplied by user
type RawData struct {
	Path      string        `json:"path"`      // Path to tiles
	Unique    bool          `json:"unique"`    // Default to false
	TileSize  int           `json:"tileSize"`  // Default to 16
	Tiles     []RawTile     `json:"tiles"`     //
	Neighbors []RawNeighbor `json:"neighbors"` //
}

// Raw information on a tile
type RawTile struct {
	Name     string  `json:"name"`     // Name used to identify the tile
	Symmetry string  `json:"symmetry"` // Default to ""
	Weight   float64 `json:"weight"`   // Default to 1
}

// Information on which tiles can be neighbors
type RawNeighbor struct {
	Left     string `json:"left"`     // Mathces Tile.Name
	LeftNum  int    `json:"leftNum"`  // Default to 0
	Right    string `json:"right"`    // Mathces Tile.Name
	RightNum int    `json:"rightNum"` // Default to 0
}

func TestNewSimpleTiledModel(t *testing.T) {
	dataFileName := "castle_data.json"
	targetFileName := "castle.png"
	periodic := true
	width := 48
	height := 48
	seed := int64(42)

	// Load data file
	dataFile, err := ioutil.ReadFile("internal/input/" + dataFileName)
	if err != nil {
		panic(err)
	}

	// Parse rawData file
	var rawData RawData
	if err := json.Unmarshal(dataFile, &rawData); err != nil {
		panic(err)
	}

	// Marshal into data settings struct
	tiles := make([]Tile, len(rawData.Tiles))
	for i, rt := range rawData.Tiles {
		imgs := make([]image.Image, 0)
		if rawData.Unique {
			i := 0
			for {
				if img, err := testutils.LoadImage("internal/input/" + rawData.Path + rt.Name + " " + strconv.Itoa(i) + ".png"); err == nil {
					imgs = append(imgs, img)
				} else {
					break
				}
				i++
			}
		} else {
			img, err := testutils.LoadImage("internal/input/" + rawData.Path + rt.Name + ".png")
			if err != nil {
				panic(err)
			}
			imgs = append(imgs, img)
		}
		tiles[i] = Tile{Name: rt.Name, Symmetry: rt.Symmetry, Weight: rt.Weight, Variants: imgs}
	}
	neighboors := make([]Neighbor, len(rawData.Neighbors))
	for i, rn := range rawData.Neighbors {
		neighboors[i] = Neighbor{Left: rn.Left, LeftNum: rn.LeftNum, Right: rn.Right, RightNum: rn.RightNum}
	}
	data := SimpleTiledData{Unique: rawData.Unique, TileSize: rawData.TileSize, Tiles: tiles, Neighbors: neighboors}

	// Create model
	var outputImg image.Image
	success := false
	model := NewSimpleTiledModel(data, width, height, periodic)
	model.SetSeed(seed)
	for !success {
		outputImg, success := model.Generate()
	}
	if !success {
		t.Log("Failed to generate image on the first try.")
		t.FailNow()
	}

	// Save output
	err = testutils.SaveImage("internal/target/"+targetFileName, outputImg)
	if err != nil {
		panic(err)
	}

	// Test that files match
	targetImg, err := testutils.LoadImage("internal/target/" + targetFileName)
	if err != nil {
		panic(err)
	}
	// areEqual := testutils.CompareImages(outputImg, targetImg)
	// if !areEqual {
	// 	t.Log("Output image is not the same as the target image.")
	// 	t.FailNow()
	// }
}
