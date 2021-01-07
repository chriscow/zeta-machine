package zeta

import (
	"image"
	"image/color"
	"math"
	"math/cmplx"
	"runtime"
	"sync"
)

const (
	minN     = 100
	maxN     = 1000000
	cabsZMax = 10000.0
	maxITs   = 5000
	maxGamma = 450.0
)

var (
	sqrt2Pi = math.Sqrt(math.Pi * 2)
	cc      = [300]uint8{0, 0, 0, 0, 0, 0, 0, 0, 255, 0, 60, 255, 0, 100, 255, 0, 125, 255, 0, 140, 255, 0, 155, 255, 0, 170, 255, 0, 180, 255, 0, 190, 255, 0, 200, 255, 0, 210, 255, 0, 220, 255, 0, 225, 255, 0, 230, 255, 0, 235, 255, 0, 238, 255, 0, 241, 255, 0, 244, 255, 0, 247, 255, 0, 250, 255, 0, 253, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 23, 255, 0, 46, 255, 0, 70, 255, 0, 93, 255, 0, 116, 255, 0, 139, 255, 0, 163, 255, 0, 186, 255, 0, 209, 255, 0, 232, 255, 0, 255, 255, 0, 255, 237, 0, 255, 218, 0, 255, 200, 0, 255, 181, 0, 255, 163, 0, 255, 146, 0, 255, 128, 0, 255, 111, 0, 255, 93, 0, 255, 76, 0, 255, 63, 0, 255, 51, 0, 255, 42, 0, 255, 39, 0, 255, 36, 0, 255, 33, 0, 255, 30, 0, 255, 27, 0, 255, 24, 0, 255, 21, 0, 255, 18, 0, 255, 15, 0, 255, 12, 0, 255, 9, 0, 255, 6, 0, 255, 3, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0}
	bCoeff  = [20]float64{
		1.0000000000000000000000000000000,
		0.0833333333333333333333333333333,
		-0.0013888888888888888888888888888,
		3.3068783068783068783068783068783e-5,
		-8.2671957671957671957671957671958e-7,
		2.0876756987868098979210090321201e-8,
		-5.2841901386874931848476822021796e-10,
		1.3382536530684678832826980975129e-11,
		-3.3896802963225828668301953912494e-13,
		8.5860620562778445641359054504256e-15,
		-2.1748686985580618730415164238659e-16,
		5.5090028283602295152026526089023e-18,
		-1.3954464685812523340707686264064e-19,
		3.5347070396294674716932299778038e-21,
		-8.9535174266605480875210207537274e-23,
		2.2679524523376830603109507388682e-24,
		-5.7447906688722024452638819876070e-26,
		1.4551724756148649018662648672713e-27,
		-3.6859949406653101781817824799086e-29,
		9.3367342570950446720325551527856e-31,
	}
	gCoeff = [15]float64{
		0.99999999999999709182,
		57.15623566586292351700,
		-59.59796035547549124800,
		14.13609797474174717400,
		-0.491913816097620199780,
		0.33994649984811888699e-4,
		0.46523628927048575665e-4,
		-0.98374475304879564677e-4,
		0.15808870322491248884e-3,
		-0.21026444172410488319e-3,
		0.21743961811521264320e-3,
		-0.16431810653676389022e-3,
		0.84418223983852743293e-4,
		-0.26190838401581408670e-4,
		0.36899182659531622704e-5,
	}
)

// Algo ...
type Algo struct {
	size   int
	stride int
	img    *image.RGBA
	wg     *sync.WaitGroup
}

// Render ...
func (a *Algo) Render(min, max complex128, img *image.RGBA) {
	procs := runtime.GOMAXPROCS(0)

	a.img = img
	a.wg = &sync.WaitGroup{}
	a.size = img.Rect.Size().X
	a.stride = a.size * a.size / procs // pixels per proc

	for start := 0; start < a.size*a.size; start += a.stride {
		a.wg.Add(1)
		go a.renderPatch(min, max, start)
		// go func(id, i int) {
		// 	// fmt.Println("running job", id, "block size", stride)
		// 	for pos := i; pos < i+stride; pos++ {
		// 		x := pos % a.Size
		// 		y := pos / a.Size
		// 		u := float64(x) / float64(a.Size)
		// 		v := float64(y) / float64(a.Size)

		// 		span := a.Max - a.Min

		// 		fmt.Println("job", id, "on pos", pos, x, y)
		// 		// delta := 1.0 / float64(a.Res)

		// 		s := a.Min + complex(real(span)*u, imag(span)*v)

		// 		var col color.RGBA
		// 		if a.LUTs == nil {
		// 			its := iterate(s, 1e-15)
		// 			col = color.RGBA{
		// 				R: cc[its*3+0],
		// 				G: cc[its*3+1],
		// 				B: cc[its*3+2],
		// 				A: 255,
		// 			}
		// 		} else {
		// 			z := zeta(s)
		// 			for l := range a.LUTs {
		// 				if c, ok := a.LUTs[l].Lookup(z); ok {
		// 					col = c
		// 					break
		// 				}
		// 			}
		// 		}

		// 		img.Set(x, y, col)
		// 	}
		// 	fmt.Println(id, "job complete")
		// 	<-sem
		// 	wg.Done()
		// }(jobid, index)

	}

	a.wg.Wait()
}

// RunSync ...
func (a *Algo) renderPatch(min, max complex128, start int) {
	defer a.wg.Done()

	span := max - min
	// ppu := float64(a.size) / real(span)

	for index := start; index < start+a.stride; index++ {
		x := index % a.size
		y := index / a.size
		u := float64(x) / float64(a.size)
		v := float64(y) / float64(a.size)
		s := min + complex(real(span)*u, imag(span)*v)

		col := color.RGBA{
			R: 255,
			G: 0,
			B: 255,
			A: 255,
		}

		// if a.LUTs == nil {
		its := iterate(s, 1e-15)
		col = color.RGBA{
			R: cc[its*3+0],
			G: cc[its*3+1],
			B: cc[its*3+2],
			A: 255,
		}
		// } else {
		// 	z := zeta(s)
		// 	for l := range a.LUTs {
		// 		if c, ok := a.LUTs[l].Lookup(z); ok {
		// 			col = c
		// 			break
		// 		}
		// 	}
		// }

		a.img.Set(x, y, col)
	}
}

func iterate(s complex128, epsilon float64) int {
	var i int
	var cabsz float64
	var diff float64 = 100

	var z complex128

	for !math.IsNaN(cabsz) && diff > epsilon && cabsz < cabsZMax && i < maxITs {
		z = zeta(s)
		diff = math.Abs(real(z) - real(s))
		cabsz = mod(z)
		i++
		s = z
	}

	if !math.IsNaN(cabsz) && cabsz >= cabsZMax {
		if real(z) < 0.0 {
			i++
		} else {
			i += 2
		}
	}
	return i
}

func zeta(s complex128) complex128 {
	var z complex128

	if real(s) < 0.0 {
		if math.Abs(imag(s)) < maxGamma {
			s = 1.0 - s
			g := gamma(s)
			z = ems(s)
			z *= g * 2.0 * cmplx.Pow(math.Pi*2.0, -s) * cmplx.Cos(math.Pi/2.0*s)
		} else {
			z = ems(s)
		}
	} else {
		z = ems(s)
	}
	return z
}

func gamma(s complex128) complex128 {
	g := complex(gCoeff[0], 0)
	s += -1
	for i := 1; i < 15; i++ {
		g += complex(gCoeff[i], 0) / (s + complex(float64(i), 0))
	}
	g *= complex(sqrt2Pi, 0)
	g *= cmplx.Pow(s+5.2421875, s+0.5)
	g *= cmplx.Exp(-5.2421875 - s)
	return g
}

func ems(s complex128) complex128 {
	N := int(cmplx.Abs(s))
	var z, t, temp complex128
	if N > maxN {
		N = maxN
	}
	if N < minN {
		N = minN
	}
	for k := 1; k < N; k++ {
		z += pow(float64(k), -s)
	}

	z += pow(float64(N), (1+0i-s)) / (s + complex(-1, 0))
	z += (0.5 + 0i) * pow(float64(N), -s)

	for k := 1; k < 20; k++ {
		t +=
			complex(bCoeff[k], 0.0) *
				pochhammer(s, (2*k)-1) *
				pow(
					float64(N),
					complex(float64(1-(2*k)), 0)-s)

		if real(t-temp) == 0.0 {
			break
		}
		temp = t
	}
	return z + t
}

func pow(a float64, c complex128) complex128 {
	re := math.Pow(a, real(c)) * math.Cos(imag(c)*math.Log(a))
	im := math.Pow(a, real(c)) * math.Sin(imag(c)*math.Log(a))

	return complex(re, im)
}

// complex modulus
func mod(a complex128) float64 {
	return math.Sqrt(real(a)*real(a) + imag(a)*imag(a))
}

func pochhammer(s complex128, n int) complex128 {
	val := 1 + 0i
	for i := 0; i < n; i++ {
		val *= (s + complex(float64(i), 0))
	}
	return val
}
