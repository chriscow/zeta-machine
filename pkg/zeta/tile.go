package zeta

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

const (
	// TotalUnits represents the number of units we render at 1 pixel per unit
	// at zoom level zero
	TotalUnits = 512

	// GlobalMin is the lower left coordinate for 1 tile at 1 pixel per unit
	GlobalMin = -256 - 256i

	// TileWidth is the width of a rendered tile in pixels
	TileWidth = 256
)

// Tile holds information for generating a single zeta tile at a particular
// zoom level
type Tile struct {
	Zoom int    `json:"zoom"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Data string `json:"data"`
}

// Render generates a single tile image using the tile's properties
func (t *Tile) Render(data []uint8, colors []color.Color) (*image.Image, error) {

	rgba := image.NewNRGBA(image.Rect(0, 0, TileWidth, TileWidth))

	for i := range data {
		x := i % TileWidth
		y := i / TileWidth
		c := data[i]
		rgba.Set(x, y, colors[c])
	}
	var img image.Image = rgba

	return &img, nil
}

// RequestToTile parses the URL parameters to get the tile arguments, then it
// constructs a *Tile instance and returns it
func RequestToTile(r *http.Request) (*Tile, error) {
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

	t := &Tile{Zoom: zoom, X: x, Y: y}

	return t, nil
}

// PPU returns the resolution of this tile in pixels per unit
func (t *Tile) PPU() int {
	return int(float64(TileWidth) / (real(t.Max() - t.Min())))
}

// IsBackground tries to determine if the tile would be rendered the background
// color and skips the calculations required so we can simply return a solid
// color
func (t *Tile) IsBackground() bool {
	// tile := tile.Tile{
	// 	Size:       size,
	// 	Zoom:       zoom,
	// 	X:          x,
	// 	Y:          y,
	// 	TotalUnits: TotalUnits,
	// 	GlobalMin:  GlobalMin,
	// }

	// max := tile.Max()
	// min := tile.Min()
	// if zoom > 3 && (real(min) >= 20) {
	// 	fmt.Println("tile:", x, y, "min:", min)
	// 	return true
	// }

	// fmt.Println("zoom:", zoom, "tile:", x, y, "extents:", min, max)
	return false
}

// Filename returns the filename for this tile
func (t *Tile) Filename() string {
	return fmt.Sprintf("%d.%d.%d.dat", t.Zoom, t.Y, t.X)
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
func (t *Tile) Min() complex128 {
	// Tile count is always even. Tile 0,0 is in the center so tile
	// numbers can be negative
	offset := float64(t.tileCount()) / 2
	x := float64(t.X) + offset
	y := float64(t.Y) + offset
	stride := t.units()

	r := real(GlobalMin) + x*stride
	i := imag(GlobalMin) + y*stride
	return complex(r, i)
}

// Max returns the upper-right coordinate in 'units' this tile renders
func (t *Tile) Max() complex128 {
	min := t.Min()
	stride := t.units()

	r := real(min) + stride
	i := imag(min) + stride
	return complex(r, i)
}

// Units is the number of 'units' this tile covers (this is not pixels)
func (t *Tile) units() float64 {
	return TotalUnits / t.tileCount()
}

func (t *Tile) String() string {
	return fmt.Sprint("x:", t.X, " y:", t.Y, " ppu:", t.PPU(), " min:", t.Min(), " max:", t.Max(), " units:", t.units())
}
