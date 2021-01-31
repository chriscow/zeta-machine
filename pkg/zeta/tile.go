package zeta

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
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
	// TileWidth is the width of a rendered tile in pixels
	TileWidth = 512
)

// Tile holds information for generating a single zeta tile at a particular
// zoom level
type Tile struct {
	Zoom  int      `json:"zoom"`
	X     int      `json:"x"`
	Y     int      `json:"y"`
	Width int      `json:"width"`
	Data  []uint16 `json:"data"`
}

// Render generates a single tile image using the tile's properties
func (t *Tile) Render(colors []color.Color) (image.Image, error) {

	rgba := image.NewNRGBA(image.Rect(0, 0, t.Width, t.Width))

	for i := range t.Data {
		x := i % t.Width
		y := i / t.Width
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

	t := &Tile{Zoom: zoom, X: x, Y: y, Width: TileWidth}

	return t, nil
}

// PPU returns the resolution of this tile in pixels per unit
func (t *Tile) PPU() int {
	return int(math.Pow(2, float64(t.Zoom)))
}

// Min returns the lower left coordinate in 'units' this tile renders
func (t *Tile) Min() complex128 {
	units := t.Units()
	r := float64(t.X) * units
	i := float64(t.Y) * units
	return complex(r, i)
}

// Max returns the upper-right coordinate in 'units' this tile renders
func (t *Tile) Max() complex128 {
	min := t.Min()
	units := t.Units()

	r := real(min) + units
	i := imag(min) + units
	return complex(r, i)
}

// Units is the number of 'units' this tile covers (this is not pixels)
func (t *Tile) Units() float64 {
	return float64(t.Width / t.PPU())
}

// RenderSolid renders a solid color tile of the given color
func (t *Tile) RenderSolid(bkg color.Color) image.Image {
	rgba := image.NewRGBA(image.Rect(0, 0, t.Width, t.Width))
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

func (t *Tile) String() string {
	return fmt.Sprint("zoom:", t.Zoom, " x:", t.X, " y:", t.Y, " ppu:", t.PPU(), " min:", t.Min(), " max:", t.Max(), " units:", t.Units(), " width:", t.Width)
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

func (t *Tile) SavePNG(colors []color.Color, fullpath string) error {

	img, err := t.Render(colors)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		log.Println("[saveTmpPNG] failed to encode: ", err)
		return err
	}

	f, err := os.Create(fullpath)
	if err != nil {
		log.Println("[saveTmpPNG] failed to open: ", fullpath, err)
		return err
	}
	defer f.Close()

	i, err := io.Copy(f, buf)
	if err != nil {
		log.Println("[saveTmpPNG] failed to copy: ", err, i)
		return err
	}

	return nil
}
