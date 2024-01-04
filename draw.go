// project: Go Interfaces
// Name : Yousef Habeeb
// Netid : yhabe2
// Date 10/27/2023

// this project uses geoometric shapes to create instances of rectangles, cricles, and triangles to draw them in a .ppm output file

package main

import (
	"errors"
	"fmt"
	"os"
)

var display Display

// color structuree and constants
type RGB struct {
	R, G, B uint8
}

type Color int

const (
	red Color = iota
	green
	blue
	yellow
	orange
	purple
	brown
	black
	white
)

// the point structure
type Point struct {
	X, Y int
}

var colorMap = map[Color]RGB{
	red:    {255, 0, 0},
	green:  {0, 255, 0},
	blue:   {0, 0, 255},
	yellow: {255, 255, 0},
	orange: {255, 164, 0},
	purple: {128, 0, 128},
	brown:  {165, 42, 42},
	black:  {0, 0, 0},
	white:  {255, 255, 255},
}

// the screen interface
type screen interface {
	initialize(maxX, maxY int)
	drawPixel(x, y int, c Color) error
	getPixel(x, y int) (Color, error)
	clearScreen()
	screenShot(filename string) error
}

// the geometry interface
type geometry interface {
	draw(scn screen) error
	shape() string
}

// the display structure and methods
type Display struct {
	maxX, maxY int
	matrix     [][]Color
}

// initializes the display with the specified maxX and maxY dimensions.
func (d *Display) initialize(maxX, maxY int) {
	d.maxX = maxX
	d.maxY = maxY
	d.matrix = make([][]Color, maxY)
	for i := range d.matrix {
		d.matrix[i] = make([]Color, maxX)
		for j := range d.matrix[i] {
			d.matrix[i][j] = white
		}
	}
}

// handles rotation of the image and checks for bounds and color validity.
func (d *Display) drawPixel(x, y int, c Color) error {

	// rotating the image to match the pdf picture
	rotatedY := x
	rotatedX := y

	if rotatedX < 0 || rotatedY < 0 || rotatedX >= d.maxX || rotatedY >= d.maxY {
		return errors.New("pixel out of bounds")
	}

	// checking if the color exists
	_, exists := colorMap[c]
	if !exists {
		return fmt.Errorf("color unknown")
	}

	d.matrix[rotatedY][rotatedX] = c

	return nil
}

// retrieves the color of the pixel at the specified (x, y) coordinates.
func (d *Display) getPixel(x, y int) (Color, error) {
	if x < 0 || y < 0 || x >= d.maxX || y >= d.maxY {
		return 0, errors.New("pixel out of bounds")
	}
	return d.matrix[y][x], nil
}

// sets all pixels on the display to the color white.
func (d *Display) clearScreen() {
	for i := range d.matrix {
		for j := range d.matrix[i] {
			d.matrix[i][j] = white
		}
	}
}

// returns the maxX and maxY dimensions of the Display.
func (d *Display) getMaxXY() (int, int) {
	return d.maxX, d.maxY
}

// takes a screenshot of the display and saves it as a .ppm image file with the provided filename.
func (d *Display) screenShot(filename string) error {
	file, err := os.Create(filename + ".ppm")
	if err != nil {
		return err
	}
	defer file.Close()

	// write the .ppm header with screen dimensions and color range
	_, err = fmt.Fprintf(file, "P3\n%d %d\n255\n", d.maxX, d.maxY)
	if err != nil {
		return err
	}

	// going through the display pixels and write their RGB values to the file.
	for y := 0; y < d.maxY; y++ {
		for x := 0; x < d.maxX; x++ {
			color, exists := colorMap[d.matrix[y][x]]
			if !exists {
				return fmt.Errorf("invalid color at pixel [%d, %d]", x, y)
			}
			_, err = fmt.Fprintf(file, "%d %d %d ", color.R, color.G, color.B)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(file)
		if err != nil {
			return err
		}
	}

	return nil
}

type Rectangle struct {
	LL Point // lower left
	UR Point // upper right
	C  Color
}

// draws a rectangle on the screen using the specified color.
func (r Rectangle) draw(scn *Display) error {

	maxX, maxY := scn.getMaxXY()
	if r.LL.X < 0 || r.LL.Y < 0 || r.UR.X >= maxX || r.UR.Y >= maxY {
		return fmt.Errorf("geometry out of bounds")
	}

	for y := r.LL.Y; y <= r.UR.Y; y++ {
		for x := r.LL.X; x <= r.UR.X; x++ {
			err := scn.drawPixel(x, y, r.C)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// returns the shape type as rectangle
func (r Rectangle) shape() string {
	return "Rectangle"
}

type Circle struct {
	CP Point // center point
	R  int   // radius
	C  Color
}

// draws a circle on the screen using the specified color.
func (c Circle) draw(scn *Display) error {
	maxX, maxY := scn.getMaxXY()
	if c.CP.X-c.R < 0 || c.CP.X+c.R >= maxX || c.CP.Y-c.R < 0 || c.CP.Y+c.R >= maxY {
		return fmt.Errorf("%s: geometry out of bounds", c.shape())
	}

	for y := -c.R; y <= c.R; y++ {
		for x := -c.R; x <= c.R; x++ {
			if x*x+y*y <= c.R*c.R {
				err := scn.drawPixel(c.CP.X+x, c.CP.Y+y, c.C)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// returns the shape type as circle
func (c Circle) shape() string {
	return "Circle"
}

type Triangle struct {
	Pt0, Pt1, Pt2 Point // edge of trianges
	C             Color
}

// this is a helper function for the draw triangle
func interpolate(y0, x0, y1, x1 int) []int {
	values := make([]int, 0, y1-y0+1)
	a := float64(x1-x0) / float64(y1-y0)
	d := float64(x0)
	for y := y0; y <= y1; y++ {
		values = append(values, int(d))
		d += a
	}
	return values
}

// draws a triangle on the screen using the specified color.
func (t Triangle) draw(scn *Display) error {
	maxX, maxY := scn.getMaxXY()
	x0, y0 := t.Pt0.X, t.Pt0.Y
	x1, y1 := t.Pt1.X, t.Pt1.Y
	x2, y2 := t.Pt2.X, t.Pt2.Y

	// make sure the triangle is within the bounds of the screen.
	if y1 < y0 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	if y2 < y0 {
		x0, x2 = x2, x0
		y0, y2 = y2, y0
	}
	if y2 < y1 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	if x0 < 0 || x1 < 0 || x2 < 0 || y0 < 0 || y1 < 0 || y2 < 0 ||
		x0 >= maxX || x1 >= maxX || x2 >= maxX || y0 >= maxY || y1 >= maxY || y2 >= maxY {
		return fmt.Errorf("geometry out of bounds")
	}

	// interpolate the edges of the triangle to fill it.
	x01 := interpolate(y0, x0, y1, x1)
	x12 := interpolate(y1, x1, y2, x2)
	x02 := interpolate(y0, x0, y2, x2)
	x012 := append(x01, x12[1:]...)

	m := len(x012) / 2
	var x_left, x_right []int
	if x02[m] < x012[m] {
		x_left, x_right = x02, x012
	} else {
		x_left, x_right = x012, x02
	}

	// the actual drawing
	for y := y0; y <= y2; y++ {
		for x := x_left[y-y0]; x <= x_right[y-y0]; x++ {
			if err := scn.drawPixel(x, y, t.C); err != nil {
				return err
			}
		}
	}
	return nil
}

// returns the shape type as triangle
func (t Triangle) shape() string {
	return "Triangle"
}
