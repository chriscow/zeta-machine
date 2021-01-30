// +build !BUILD_CUDA

package seed

import "math/rand"

// Generate tile data via call to cuda zeta machine library
func Generate(p Patch) []uint32 {
	data := make([]uint32, PatchWidth*PatchWidth)
	for i := range data {
		data[i] = rand.Uint32()
	}
	return data
}
