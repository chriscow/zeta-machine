// +build !BUILD_CUDA

package cuda

import "zetamachine/pkg/zeta"

// Generate tile data via call to cuda zeta machine library
func Generate(p zeta.Patch) []uint32 {
	data := make([]uint32, 0)
	return data
}
