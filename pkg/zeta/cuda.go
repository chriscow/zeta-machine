// +build BUILD_CUDA

package zeta

import (
	"context"
	"log"
	"time"
)

// compile the cuda code from the root workspace folder with:
//
//	nvcc --ptxas-options=-v --compiler-options '-fPIC' -o ./cmd/cuda/libzm.so --shared ./cuda/zeta.cu ./cuda/main.cu
//
// compile the go code with:
//
//	go build -tags BUILD_CUDA -o build/cuda ./cmd/cuda/.
//

/*
void generate(double minR, double maxR, double minI, double maxI, unsigned int size, unsigned int* data);
#cgo LDFLAGS: -L. -L./ -lzm
*/
import "C"

// Generate tile data via call to cuda zeta machine library
func (t *Tile) ComputeRequest(ctx context.Context) {
	start := time.Now()
	buf := make([]C.uint, p.Width*p.Width)
	C.generate(C.double(p.Min[0]), C.double(p.Max[0]), C.double(p.Min[1]), C.double(p.Max[1]), C.uint(p.Width), &buf[0])

	t.Data = make([]uint16, len(buf))
	for i := range buf {
		t.Data[i] = uint16(buf[i])
	}
	log.Println("[tile] compute complete in ", time.Since(start), t)
}
