package zeta

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"time"

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
	Zoom      int    `json:"zoom"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Size      int    `json:"size"`
	Data      string `json:"data"`
	upsampled bool
}

// Render generates a single tile image using the tile's properties
func (t *Tile) Render(data []uint8, colors []color.Color) (image.Image, error) {

	rgba := image.NewNRGBA(image.Rect(0, 0, TileWidth, TileWidth))

	for i := range data {
		x := i % TileWidth
		y := i / TileWidth
		c := data[i]
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
	data := algo.Compute(ctx, tile.Min(), tile.Max(), luts)
	tile.Data = base64.StdEncoding.EncodeToString(data)
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

// ShouldGenerate checks for the existence of the tile and also for the temp file
// created when a tile is requested.
// If the tile data already exists, ShouldGenerate returns true and makes sure
// there is no temporary .req file still hanging around.
//
// If the tile data doesn't exist, we check for a temporary .req file in its place
// to see if it has recently been requested to avoid re-requesting generation.
//
// If we find a .req file AND it is less than `maxAgeMin` minuites old we return false. It is
// already in the queue to be generated. If it is older than 24 hours, we delete
// the temporary .req file and re-request generation.
//
func (t *Tile) ShouldGenerate(maxAge time.Duration) (bool, error) {
	// the tile tile already exists, return not ok
	existing, err := t.Exists()
	if existing != nil {
		// Since the tile data already exists, make sure there isn't a
		// left over .req temp file hanging around
		t.Received()
		return false, err
	}

	// folder should exist no matter what
	fpath := t.Path()
	if err := os.MkdirAll(fpath, os.ModeDir|os.ModePerm); err != nil {
		log.Println("[ShouldGenerate] failed to create: ", t.Path())
		log.Println("[ShouldGenerate] failed: ", err)
		return false, err
	}

	// see if we already have this tile flagged with a temp file
	fname := path.Join(fpath, t.Filename()+".req")
	info, err := os.Stat(fname)
	if err != nil {
		return false, err
	}

	if info != nil {
		// check its age
		age := time.Since(info.ModTime())
		if age > maxAge {
			// file is old. remove it
			os.Remove(fpath)
		} else {
			// file exists and it has not timed out
			return false, nil
		}
	}

	// the .req file either didn't exist or is older than 24 hours and was deleted above
	// tell the client to go ahead and re-request it
	//
	// we'll create a temporary `.req` file as a flag to indicate we have
	// already requested the file and it is pending, in case someone comes asking again.
	f, err := os.Create(fpath)
	if err != nil {
		log.Println("[ShouldGenerate] failed to create: ", fpath)
		log.Println("[ShouldGenerate] failed: ", err)
		return false, err
	}
	defer f.Close()

	return true, nil
}

// Received is called to remove the temp file that was created when this tile
// was requested to be generated
func (t *Tile) Received() {
	fname := t.Filename() + ".req"
	fpath := path.Join(t.Path(), fname)
	if err := os.Remove(fpath); err != nil {
		log.Println("[tile::Received] could not remove tmp file: ", err)
		return
	}
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

	data, err := base64.StdEncoding.DecodeString(t.Data)
	if err != nil {
		return err
	}

	// does not return an error if the path exists. creates the path recusively
	if err := os.MkdirAll(fpath, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bytes.NewBuffer(data)
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	return nil
}

// NoResample is the error string returned if sampling fails
const NoResample = "Cannot resample tile. Suitable reference tile(s) don't exist"

// Downsample looks for a higher resolution tile and samples the iteration data
// to create a scaled down version.
//
// If we cannot downsample, an error is returned.
func (t *Tile) Downsample() (*Tile, error) {
	return nil, errors.New(NoResample)
}

// Upsample looks for a lower resolution tile and scales it up for a placeholder
// tile to return to the user while the full resolution tile gets rendered.
//
// If we cannot upsample, an error is returned.
func (t *Tile) Upsample() (*Tile, error) {
	return nil, errors.New(NoResample)
}
