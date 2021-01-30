// +build !BUILD_CUDA

package seed

// Generates tile data via call to cuda zeta machine library
func (p *Patch) Generate() {
	p.Data = make([]uint16, PatchWidth*PatchWidth)
	for i := range p.Data {
		p.Data[i] = uint16(i % 5000) // max iterations
	}
}
