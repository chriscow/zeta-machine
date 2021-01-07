package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"zetamachine/pkg/generator"
	"zetamachine/pkg/zeta"
)

var (
	zoom, mapRange     *int
	xTileMin, xTileMax *int
	mapLow             *complex128
)

func findHorizRange() {

}

func generate(zoom, xmin, xmax int) {
	luts := zeta.LoadLUTs()
	wg := &sync.WaitGroup{}
	procs := runtime.GOMAXPROCS(0)
	fmt.Println("num procs", procs)

	z := strconv.Itoa(zoom)

	numtiles := math.Pow(2, float64(zoom))

	if numtiles < float64(xmax) {
		xmax = int(numtiles) - 1
	}

	sem := make(chan bool, procs)

	cwd, _ := os.Getwd()
	zoomPath := path.Join(cwd, "public/tiles", z)
	if err := createFolder(zoomPath); err != nil {
		log.Fatal(zoomPath, " ", err)
	}

	for x := xmin; x <= xmax; x++ {
		xPath := path.Join(zoomPath, strconv.Itoa(x))
		if err := createFolder(xPath); err != nil {
			log.Fatal(xPath, " ", err)
		}

		for y := 0.0; y < numtiles; y++ {
			sem <- true
			wg.Add(1)
			go makeTile(luts, zoom, -30, 60, float64(x), y, wg, sem)
		}
	}

	log.Println("waiting for finish")
	wg.Wait()
}

func makeTile(luts []*zeta.LUT, zoom int, min, span, x, y float64, wg *sync.WaitGroup, sem <-chan bool) {
	defer func() { <-sem }()
	defer wg.Done()

	cwd, _ := os.Getwd()

	img, err := zeta.MakeTile(luts, zoom, min, span, x, y)
	if err != nil {
		log.Fatal(err)
	}

	z := strconv.Itoa(zoom)

	fname := fmt.Sprintf("%d-%d-%d.png", int(zoom), int(x), int(y))
	fpath := path.Join(cwd, "public/tiles", z, strconv.Itoa(int(x)), fname)
	if _, err := os.Stat(fpath); err == nil {
		log.Fatal(err)
	}

	if err := generator.Save(img, fpath); err != nil {
		log.Fatal(err)
	}
}

func createFolder(path string) error {
	exists, err := pathExists(path)
	if err != nil {
		return err
	}

	if !exists {
		err := os.Mkdir(path, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// exists returns whether the given file or directory exists or not
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
