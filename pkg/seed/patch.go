package seed

import (
	"fmt"
	"math"
	"zetamachine/pkg/palette"
	"zetamachine/pkg/zeta"
)

const (
	PatchWidth = zeta.TileWidth * 4
)

// Patch is similar to a tile but instead is a larger patch of data that will
// be split into tiles.  The purpose is to have powerful GPUs calculate much
// larger areas than just a single tile in one go, then split them up into
// a suitable size for display on the web
type Patch struct {
	ID    uint64    `json:"id"`
	Zoom  uint8     `json:"zoom"`
	X     int       `json:"x"`
	Y     int       `json:"y"`
	Width int       `json:"width"`
	Min   []float64 `json:"min"`
	Max   []float64 `json:"max"`
	Data  []uint16  `json:"data"`
}

// NewPatch ...
func NewPatch(id uint64, zoom uint8, min, max complex128, x, y, width int) *Patch {
	p := &Patch{
		ID:    id,
		Zoom:  zoom,
		X:     x,
		Y:     y,
		Width: width,
	}
	p.SetMin(min)
	p.SetMax(max)

	return p
}

func (p *Patch) String() string {
	return fmt.Sprint("id:", p.ID, " zoom:", p.Zoom, " min:", p.Min, " max:", p.Max)
}

// SetMin ...
func (p *Patch) SetMin(min complex128) {
	p.Min = make([]float64, 2)
	p.Min[0] = real(min)
	p.Min[1] = imag(min)
}

// SetMax ...
func (p *Patch) SetMax(max complex128) {
	p.Max = make([]float64, 2)
	p.Max[0] = real(max)
	p.Max[1] = imag(max)
}

// GetMin ...
func (p *Patch) GetMin() complex128 {
	return complex(p.Min[0], p.Min[1])
}

// GetMax ...
func (p *Patch) GetMax() complex128 {
	return complex(p.Max[0], p.Max[1])
}

// SavePNG ...
func (p *Patch) SavePNG() error {
	ppu := math.Pow(2, float64(p.Zoom))
	units := p.Max[0] - p.Min[0]

	tile := zeta.Tile{
		Zoom:  int(p.Zoom),
		X:     p.X,
		Y:     p.Y,
		Width: int(units * ppu),
		Data:  p.Data,
	}

	return tile.SavePNG(palette.DefaultPalette)
}

// Split splits the patch data into individual tiles
func (p *Patch) Split() ([]*zeta.Tile, error) {

	tiles := make([]*zeta.Tile, 4*4)

	// Patch data encompasses 8 tiles, each tile width and height are the same
	//			 ____ ____ ____ ____
	//			|    |    |    |    |
	//			|____|____|____|____|
	//			|    |    |    |    |
	//			|____|____|____|____|
	//			|    |    |    |    |
	//			|____|____|____|____|
	//			|    |    |    |    |
	//			|____|____|____|____|
	//
	for i := range tiles {
		tiles[i] = &zeta.Tile{
			Zoom:  int(p.Zoom),
			Data:  make([]uint16, zeta.TileWidth*zeta.TileWidth),
			Width: zeta.TileWidth,
		}

		for row := 0; row < zeta.TileWidth; row++ {
			start := i*zeta.TileWidth + row*p.Width

			// copy one row of data from the patch to the tile
			copy(tiles[i].Data[row*zeta.TileWidth:row*zeta.TileWidth+zeta.TileWidth],
				p.Data[start:start+zeta.TileWidth])
		}
		/*
			unitsPerPatch = patch.Max[0] - patch.Min[1]
			unitsPerTile = unitsPerPatch / 4

			patch.X = unitsPerPatch % patch.Max[0]
			patch.Y = unitsPerPatch % patch.Max[1]

			tile.X = patch.X * 4 + i % 4
			tile.Y = patch.Y * 4 + i / 4
			units := p.Max[0] - p.Min[1]
			unitsPerTile := units / 4 // 4 tiles across, 4 tiles down
			pixelsPerUnit := int(math.Pow(2, float64(p.Zoom)))
		*/

		tiles[i].X = p.X*4 + i%4
		tiles[i].Y = p.Y*4 + i/4

		// x = -4 + i % 4
		// y = 4 -

		// tile[i].X = int(math.Floor(patch.Min[0] + (i)/float64(PatchWidth)))
		// tile[i].Y = int(math.Floor(patch.Min[1] / float64(PatchWidth)))
	}

	return tiles, nil
}
