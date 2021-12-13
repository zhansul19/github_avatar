package main

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/llgcode/draw2d/draw2dimg"
)

type Identity struct {
	Name      string
	Hash      [16]byte
	Color     [3]byte
	Grid      []byte
	GridPoint []GridPoint
	pixelMap  []DrawingPoint
}
type GridPoint struct {
	value byte
	index int
}
type Point struct {
	x, y int
}

type DrawingPoint struct {
	topLeft     Point
	bottomRight Point
}

func hashInput(input []byte) Identity {
	hashsum := md5.Sum(input)
	return Identity{
		Name: string(input),
		Hash: hashsum,
	}
}
func Colour(identity Identity) Identity {
	rgb := [3]byte{}
	copy(rgb[:], identity.Hash[:3])
	identity.Color = rgb
	return identity
}

func buildGrid(identity Identity) Identity {
	grid := []byte{}

	for i := 0; i < len(identity.Hash) && i+3 <= len(identity.Hash)-1; i += 3 {
		chunk := make([]byte, 5)
		copy(chunk, identity.Hash[i:i+3])
		chunk[3] = chunk[1]
		chunk[4] = chunk[0]
		grid = append(grid, chunk...)
	}
	identity.Grid = grid

	return identity
}
func FilterOddSquares(identity Identity) Identity {
	grid := []GridPoint{}
	for i, code := range identity.Grid {
		if code%2 == 1 {
			point := GridPoint{
				value: code,
				index: i,
			}
			grid = append(grid, point)
		}
	}
	identity.GridPoint = grid
	return identity
}

func buildPixelMap(iden Identity) Identity {
	drawingPoints := []DrawingPoint{}
	pixelFunc := func(p GridPoint) DrawingPoint {
		horizontal := (p.index % 5) * 50
		vertical := (p.index / 5) * 50
		topLeft := Point{horizontal, vertical}
		bottomRight := Point{horizontal + 50, vertical + 50}

		return DrawingPoint{
			topLeft,
			bottomRight,
		}
	}

	for _, gridPoint := range iden.GridPoint {
		drawingPoints = append(drawingPoints, pixelFunc(gridPoint))
	}
	iden.pixelMap = drawingPoints
	return iden
}
func rect(img *image.RGBA, col color.Color, x1, y1, x2, y2 float64) {
	gc := draw2dimg.NewGraphicContext(img) // Prepare new image context
	gc.SetFillColor(col)                   // set the color
	gc.MoveTo(x1, y1)                      // move to the topleft in the image
	// Draw the lines for the dimensions
	gc.LineTo(x1, y1)
	gc.LineTo(x1, y2)
	gc.MoveTo(x2, y1) // move to the right in the image
	// Draw the lines for the dimensions
	gc.LineTo(x2, y1)
	gc.LineTo(x2, y2)
	// Set the linewidth to zero
	gc.SetLineWidth(0)
	// Fill the stroke so the rectangle will be filled
	gc.FillStroke()
}

func drawRectangle(identity Identity) error {
	// We create our default image containing a 250x250 rectangle
	var img = image.NewRGBA(image.Rect(0, 0, 250, 250))
	// We retrieve the color from the color property on the identicon
	col := color.RGBA{identity.Color[0], identity.Color[1], identity.Color[2], 255}

	// Loop over the pixelmap and call the rect function with the img, color and the dimensions
	for _, pixel := range identity.pixelMap {
		rect(
			img,
			col,
			float64(pixel.topLeft.x),
			float64(pixel.topLeft.y),
			float64(pixel.bottomRight.x),
			float64(pixel.bottomRight.y),
		)
	}
	// Finally save the image to disk
	return draw2dimg.SaveToPngFile(identity.Name+".png", img)
}

type Apply func(Identity) Identity

func pipe(identicon Identity, funcs ...Apply) Identity {
	for _, applyer := range funcs {
		identicon = applyer(identicon)
	}
	return identicon
}

func main() {
	name := "zhansultan"
	fmt.Println(name)

	data := []byte(name)
	identicon := hashInput(data)

	// Pass in the identicon, call the methods which you want to transform
	identicon = pipe(identicon, Colour, buildGrid, FilterOddSquares, buildPixelMap)

	// we can use the identicon to insert to our drawRectangle function
	if err := drawRectangle(identicon); err != nil {
		log.Fatalln(err)
	}

}
