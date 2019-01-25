# Wave Function Collapse
Go port of the Wave Function Collapse algorithm originally created by ExUtumno [https://github.com/mxgmn/WaveFunctionCollapse](https://github.com/mxgmn/WaveFunctionCollapse)

The Wave Function Collapse algorithm is a random pattern generator based on methods found in quantum physics. A sample input of constraints is fed in along with desired output criteria, i.e. width, height, periodic. The algorithm begins by analyzing the input constraints and building a list of rules about how the patterns can fit together. The output is then created in a "superposed" state, in which each slot of the output contains all possible patterns. During the first iteration, a slot is selected semi-randomly and "observed", i.e. narrowed down to one randomly chosen pattern. The implications of this selection are then propagated to the slot's neighbors recursively, eliminating patterns that cannot exist in those slots given the constraints. This cycle of observe and propagate is then repeated until all slots have one patten choosen, or there is a contradiction in which a slot has zero possible patterns. 

![Input Image](/internal/input/flowers.png?raw=true "Input Image")
![Output Image](/internal/snapshots/flowers.png?raw=true "Output Image")

![Input Images](/internal/input/castle/wallroad.png?raw=true "Input Images")
![Input Images](/internal/input/castle/bridge.png?raw=true "Input Images")
![Input Images](/internal/input/castle/wall.png?raw=true "Input Images")
![Output Image](/internal/snapshots/castle.png?raw=true "Output Image")

## Installation
The project can be fetched via Go's CLI with the following command.
```
go get github.com/shawnridgeway/wavefunctioncollapse
```
Otherwise, the project can be imported with the following line.
```
import "github.com/shawnridgeway/wavefunctioncollapse"
```

## Usage
There are two models included in this project: 

- Overlapping Model: this algorithm takes a sample image as input and uses its patterns to create a randomized texture as output.

- Simple Tiled Model: this algorithm takes in a list of tiles and constraints and produces a randomized permutation.

Each of the models has its own constructor, but all other methods are common to the two models.

### Constructors
#### Overlapping Model
```
NewOverlappingModel(inputImage image.Image, n, width, height int, periodicInput, periodicOutput bool, symmetry int, ground bool) *OverlappingModel
```
Accepts:
- `inputImage image.Image`: the sample image which will be used to extract patterns for the output.
- `n int`: the size of the patterns that the algorithm should extract. The algorithm will extract `n` by `n` square patterns from the input image to be used in constructing the output. Larger values will enable the algorithm to capture larger features in the input image, at a cost to the performance.
- `width int`: width in pixels of the output image
- `height int`: height in pixels of the output image
- `periodicInput bool`: true if the algorithm should consider the input to be repeating. This enables pattern detection to connect top-bottom and left-right pixels to form additional patterns.
- `periodicOutput bool`: true if the output should be repeatable. This means that continuity is preserved across top-bottom and left-right borders and that the image will appear seamless when tiled.
- `symmetry int`: the number of axies of symetry to consider when constructing the output. A larger value implies more reflections and rotations of the extracted patterns. Acceptable values are integers from `1` throught `8`.
- `ground bool`: true if the algorithm should look for a repeating ground pattern. This has the effect of making the bottom tiles appear only on the bottom of the output image.

Returns:
- `*OverlappingModel`: a pointer to the newly constructed model

### Simple Tiled Model
```
NewSimpleTiledModel(data SimpleTiledData, width, height int, periodic bool) *SimpleTiledModel
```
Accepts:
- `data SimpleTiledData`: data structure of tiles and constraints to be used.
	- `Unique bool`: true if the tile set contains variations of each tile.
	- `TileSize int`: the width and height in pixels of each tile.
	- `Tiles []Tile`: list of tiles to be used in the generation.
		- `Name string`: identifying name of the tile.
		- `Symetry string`: axies of symetry. Acceptable values are `"L"`, `"T"`, `"I"`, `"\\"` or `"X"`.
		- `Weight float64`: the desired frequency of this tile in the output. Values less than `1` will appear less often while those above `1` will appear more often. Note that `0` is not acceptable and will be converted to `1`. If you wish to turn a tile off, please remove it from the list.
		- `Variants []image.Image`: list of images that can be used when rendering this tile.
	- `Neighbors []Neighbor`: list of tile neighbor constraints. Defines which tiles can apper next to eachother.
		- `Left string`: name of the first tile in the pair
		- `LeftNum int`: variation number of the first tile in the pair
		- `Right string`: name of the second tile in the pair
		- `RightNum int`: variation number of the second tile in the pair
- `width int`: width in tiles of the output image
- `height int`: height in tiles of the output image
- `periodic bool`: true if the output should be repeatable. This means that continuity is preserved across top-bottom and left-right borders and that the image will appear seamless when tiled.

Returns:
- `*SimpleTiledModel`: a pointer to the newly constructed model.

### Common Methods
#### Generate
Run the algorithm until success or contradiction.
```
(model *Model) Generate() (image.Image, bool)
```

Returns:
- `image.Image`: the output image.
- `bool`: true if the generation was successful, false if a contradiction was encountered.

#### Iterate
Run the algorithm through `iterations` number of generations, stopping at success or contradiction.
```
(model *Model) Iterate(iterations int) (image.Image, bool, bool)
```
Accepts:
- `iterations int`: the number of generations to iterate over.

Returns:
- `image.Image`: the output image
- `bool`: true if the algorithm finished and cannot iterate further.
- `bool`: true if the generation was successful, false if a contradiction was encountered.

#### Render
Returns an `image.Image` of the output at its current state. This is often not necessary since both `Generate` and `Iterate` return the output image as well.
```
(model *Model) Render() image.Image
```

Returns:
- `image.Image`: the output image at its current state.

#### IsGenerationSuccessful
Returns true if the generation is finished and successful, i.e. has no contradiction.
```
(baseModel *Model) IsGenerationSuccessful() bool
```

Returns:
- `bool`: true if the generation is finished and successful, i.e. has no contradiction.

#### Clear
Clear the internal state of the algorithm for a new generation. Only necessary to call when using `Iterate` and the algorithm has not finished.
```
(model *Model) Clear()
```

#### SetSeed
Sets a stable seed for the random number generator. Unless this method is called, the model will use a seed based on the current time each time the model is reset. This method is mostly useful for creating a reproducable tests.
```
(baseModel *Model) SetSeed(seed int64)
```

Accepts: 
- `seed int64`: seed value to feed to the random number generator.

## Examples
Overlapping Model
```
// Create a new model
model := wavefunctioncollapse.NewOverlappingModel(inputImg, 3, 48, 48, true, true, 2, true)

// Run the algorithm until success or contradiction
outputImg, success := model.Generate() 

// Run the algorithm i times, stopping at success or contradiction
outputImg, finished, success := model.Iterate(i) 
```

Simple Tiled Model
```
// Create a new model
model := wavefunctioncollapse.NewSimpleTiledModel(data, 20, 20, false)

// Run the algorithm until success or contradiction
outputImg, success := model.Generate() 

// Run the algorithm i times, stopping at success or contradiction
outputImg, finished, success := model.Iterate(i) 
```

## Credits
This project was based off the JavaScript version by kchapelier ([https://github.com/kchapelier/wavefunctioncollapse](https://github.com/kchapelier/wavefunctioncollapse)) and the original version by ExUtumno ([https://github.com/mxgmn/WaveFunctionCollapse](https://github.com/mxgmn/WaveFunctionCollapse))