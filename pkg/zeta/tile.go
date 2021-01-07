package zeta

import (
	"fmt"
	"image"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

const (
	// at 1 pixel per unit how many units do we get and what are the coords
	totalUnits = 512
	globalMin  = -256 - 256i
	tileWidth  = 256 // in pixels
)

// Tile holds information for generating a single zeta tile at a particular
// zoom level
type Tile struct {
	// Size       int        // size in pixels
	Zoom int // zoom level
	X, Y int // tile number
	// TotalUnits float64    // world extents in 'units' for the entire thing
	// GlobalMin  complex128 // lower left coordinate where the world starts
	img *image.RGBA
}

// NewTile generates a single tile at the given zoom and x / y tile number
func NewTile(zoom, x, y int) (*image.Image, error) {

	// fmt.Println("generating ", t.Zoom, ".", t.X, ".", t.Y, ".png, ", t)
	t := &Tile{Zoom: zoom, X: x, Y: y}

	rgba := image.NewRGBA(image.Rect(0, 0, tileWidth, tileWidth))

	algo := &Algo{}
	algo.Render(t.min(), t.max(), rgba)

	var img image.Image = rgba
	return &img, nil
}

func ParseTileArgs(r *http.Request) (*Tile, error) {
	zoom, err := strconv.Atoi(chi.URLParam(r, "zoom"))
	if err != nil {
		return nil, err
	}

	x, err := strconv.Atoi(chi.URLParam(r, "x"))
	if err != nil {
		return nil, err
	}

	y, err := strconv.Atoi(chi.URLParam(r, "y"))
	if err != nil {
		return nil, err
	}

	return &Tile{Zoom: zoom, X: x, Y: y}, nil
}

// Filename returns the filename for this tile
func (t *Tile) Filename() string {
	return fmt.Sprintf("%d.%d.%d.png", t.Zoom, t.Y, t.X)
}

// Path returns the full relative path to the file
func (t *Tile) Path() string {
	return fmt.Sprintf("public/tiles/%d/%d", t.Zoom, t.Y)
}

// tileCount is the number of tiles in each direction required to render
// at the set zoom level
func (t *Tile) tileCount() float64 {
	return math.Pow(2, float64(t.Zoom+1))
}

// Min returns the lower left coordinate in 'units' this tile renders
func (t *Tile) min() complex128 {
	// Tile count is always even. Tile 0,0 is in the center so tile
	// numbers can be negative
	offset := float64(t.tileCount()) / 2
	x := float64(t.X) + offset
	y := float64(t.Y) + offset
	stride := t.units()

	r := real(globalMin) + x*stride
	i := imag(globalMin) + y*stride
	return complex(r, i)
}

// Max returns the upper-right coordinate in 'units' this tile renders
func (t *Tile) max() complex128 {
	min := t.min()
	stride := t.units()

	r := real(min) + stride
	i := imag(min) + stride
	return complex(r, i)
}

// Units is the number of 'units' this tile covers (this is not pixels)
func (t *Tile) units() float64 {
	return totalUnits / t.tileCount()
}

func (t *Tile) String() string {
	return fmt.Sprint("x:", t.X, " y:", t.Y, " min:", t.min(), " max:", t.max(), " units:", t.units())
}
