// +build BUILD_CUDA

package seed

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
func Generate(p Patch) []uint32 {
	buf := make([]C.uint, p.Size*p.Size)

	C.generate(C.double(p.Min[0]), C.double(p.Max[0]), C.double(p.Min[1]), C.double(p.Max[1]), C.uint(p.Size), &buf[0])

	data := make([]uint32, len(buf))
	for i := range buf {
		data[i] = uint32(buf[i])
	}

	return data
}
