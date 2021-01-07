package zeta

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"golang.org/x/image/bmp"
)

// LUT ...
type LUT struct {
	ID            int
	FName         string
	Min, Max      complex128
	Res           int
	width, height int
	img           image.Image
}

// LoadAll loads all look up tables and returns them in a sorted array
func LoadLUTs() []*LUT {
	luts := make([]*LUT, 6)
	luts[0] = &LUT{ID: 0, FName: "CL100000.bmp", Min: .975 - .025i, Max: 1.025 + .025i, Res: 100000}
	luts[1] = &LUT{ID: 1, FName: "CL010000.bmp", Min: .75 - .25i, Max: 1.25 + .25i, Res: 10000}
	luts[2] = &LUT{ID: 2, FName: "CL001000.bmp", Min: -1.5 - 2.5i, Max: 3.5 + 2.5i, Res: 1000}
	luts[3] = &LUT{ID: 3, FName: "CL000100.bmp", Min: -24 - 25i, Max: 26 + 25i, Res: 100}
	luts[4] = &LUT{ID: 4, FName: "CL000010.bmp", Min: -249 - 250i, Max: 251 + 250i, Res: 10}
	luts[5] = &LUT{ID: 5, FName: "CL000001.bmp", Min: -2499 - 2500i, Max: 2501 + 2500i, Res: 1}
	fmt.Print("Loading LUTs...")
	start := time.Now()

	wg := &sync.WaitGroup{}
	wg.Add(len(luts))
	for i := range luts {
		go func(id int) {
			l := luts[id]
			if err := l.Load(); err != nil {
				log.Fatal("error loading luts", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	sort.Slice(luts, func(i, j int) bool { return luts[i].ID < luts[j].ID })
	fmt.Println("in", time.Since(start).Milliseconds(), "ms")

	return luts
}

// Load ...
func (l *LUT) Load() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fpath := path.Join(cwd, "cuda/lut", l.FName)
	log.Println("loading ", fpath)
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}

	img, err := bmp.Decode(f)
	if err != nil {
		return err
	}

	l.img = img
	l.width = img.Bounds().Max.X - img.Bounds().Min.X
	l.height = img.Bounds().Max.Y - img.Bounds().Min.Y

	return nil
}

// Lookup ...
func (l *LUT) Lookup(z complex128) (color.RGBA, bool) {
	rng := l.Max - l.Min
	u := (real(z) - real(l.Min)) / real(rng)
	v := (imag(z) - imag(l.Min)) / imag(rng)

	x := int(u * float64(l.width))
	y := int(v * float64(l.height))

	var col color.RGBA
	ok := x >= 0 && x < l.width && y >= 0 && y < l.height
	if ok {
		col = l.img.At(x, y).(color.RGBA)
		col.A = 255
	}

	return col, ok
}
