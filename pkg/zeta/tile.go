package zeta

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
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
	Zoom      int      `json:"zoom"`
	X         int      `json:"x"`
	Y         int      `json:"y"`
	Size      int      `json:"size"`
	Data      []uint16 `json:"data"`
	upsampled bool
}

// Render generates a single tile image using the tile's properties
func (t *Tile) Render(colors []color.Color) (image.Image, error) {

	rgba := image.NewNRGBA(image.Rect(0, 0, TileWidth, TileWidth))

	for i := range t.Data {
		x := i % TileWidth
		y := i / TileWidth
		c := t.Data[i]
		rgba.Set(x, y, colors[c])
	}
	var img image.Image = rgba

	return img, nil
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

// ComputeRequest takes a JSON serialized tile, unmarshals it, computes the
// iteration data, base64 encodes the data and marshals it all back to JSON
func ComputeRequest(ctx context.Context, b []byte, luts []*LUT) ([]byte, error) {
	tile := &Tile{}
	if err := json.Unmarshal(b, tile); err != nil {
		return nil, err
	}

	log.Println("[tile] computing: ", tile)

	algo := &Algo{}
	tile.Data = algo.Compute(ctx, tile.Min(), tile.Max(), TileWidth*TileWidth)
	log.Println("[tile] compute complete: ", tile)

	return json.Marshal(tile)
}

// PPU returns the resolution of this tile in pixels per unit
func (t *Tile) PPU() int {
	return int(float64(TileWidth) / (real(t.Max() - t.Min())))
}

// Min returns the lower left coordinate in 'units' this tile renders
func (t *Tile) Min() complex128 {
	// Tile count is always even. Tile 0,0 is in the center so tile
	// numbers can be negative
	offset := float64(t.tileCount()) / 2
	x := float64(t.X) + offset
	y := float64(t.Y) + offset
	stride := t.Units()

	r := real(GlobalMin) + x*stride
	i := imag(GlobalMin) + y*stride
	return complex(r, i)
}

// Max returns the upper-right coordinate in 'units' this tile renders
func (t *Tile) Max() complex128 {
	min := t.Min()
	stride := t.Units()

	r := real(min) + stride
	i := imag(min) + stride
	return complex(r, i)
}

// Units is the number of 'units' this tile covers (this is not pixels)
func (t *Tile) Units() float64 {
	return TotalUnits / t.tileCount()
}

// RenderSolid renders a solid color tile of the given color
func (t *Tile) RenderSolid(bkg color.Color) image.Image {
	rgba := image.NewRGBA(image.Rect(0, 0, TileWidth, TileWidth))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{bkg}, image.ZP, draw.Src)

	var img image.Image = rgba
	return img
}

// Filename returns the filename for this tile
func (t *Tile) Filename() string {
	return fmt.Sprintf("%d.%d.%d.dat", t.Zoom, t.Y, t.X)
}

// Path returns the full relative path to the file
func (t *Tile) Path() string {
	tilePath := os.Getenv("ZETA_TILE_PATH")
	return path.Join(tilePath, fmt.Sprintf("%d/%d", t.Zoom, t.Y))
}

// Exists checks if the tile is already on the local disk
func (t *Tile) Exists() (os.FileInfo, error) {
	// see if we already have this tile
	fname := path.Join(t.Path(), t.Filename())
	return os.Stat(fname)
}

// tileCount is the number of tiles in each direction required to render
// at the set zoom level
func (t *Tile) tileCount() float64 {
	return math.Pow(2, float64(t.Zoom+1))
}

func (t *Tile) String() string {
	return fmt.Sprint("zoom:", t.Zoom, " x:", t.X, " y:", t.Y, " ppu:", t.PPU(), " min:", t.Min(), " max:", t.Max(), " units:", t.Units(), " width:", TileWidth)
}

// Save saves the binary iteration data from a tile
func (t *Tile) Save() error {
	fpath := t.Path()
	fname := path.Join(fpath, t.Filename())

	// does not return an error if the path exists. creates the path recusively
	if err := os.MkdirAll(fpath, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(t.Data); err != nil {
		return err
	}

	_, err = io.Copy(f, &buf)
	if err != nil {
		return err
	}

	return nil
}

// Load ...
func (t *Tile) Load() error {
	fpath := t.Path()
	fname := path.Join(fpath, t.Filename())

	if _, err := os.Stat(fname); err == nil {
		f, err := os.Open(fname)
		if err != nil {
			log.Println("Failed to open data file: ", err)
			return err
		}
		defer f.Close()

		dec := gob.NewDecoder(f)
		if err := dec.Decode(&t.Data); err != nil {
			log.Println("Failed to decode data file:", err)
			return err
		}
	} else {
		err = errors.New("Tile not found")
	}

	return nil
}
