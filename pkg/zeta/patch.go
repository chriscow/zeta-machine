package zeta

const PatchSize = TileWidth * 4

// Patch is similar to a tile but instead is a larger patch of data that will
// be split into tiles.  The purpose is to have powerful GPUs calculate much
// larger areas than just a single tile in one go, then split them up into
// a suitable size for display on the web
type Patch struct {
	ID   int       `json:"id"`
	Min  []float64 `json:"min"`
	Max  []float64 `json:"max"`
	Size uint      `json:"size"`
	Data []uint32  `json:"data"`
}

func NewPatch(min, max complex128) Patch {
	p := Patch{Size: PatchSize}
	p.SetMin(min)
	p.SetMax(max)

	return p
}

func (p Patch) SetMin(min complex128) {
	p.Min = make([]float64, 2)
	p.Min[0] = real(min)
	p.Min[1] = imag(min)
}

func (p Patch) SetMax(max complex128) {
	p.Max = make([]float64, 2)
	p.Max[0] = real(max)
	p.Max[1] = imag(max)
}

func (p Patch) GetMin() complex128 {
	return complex(p.Min[0], p.Min[1])
}

func (p Patch) GetMax() complex128 {
	return complex(p.Max[0], p.Max[1])
}
