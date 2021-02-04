package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"
	"time"
	"zetamachine/pkg/palette"
	"zetamachine/pkg/zeta"

	"github.com/briandowns/spinner"
	"github.com/joho/godotenv"
)

var (
	zoom       int
	tileCount  float64
	host, port string
)

func main() {
	if err := checkEnv(); err != nil {
		log.Fatal(err)
	}

	zoom := flag.Int("zoom", 0, "zoom to render")
	minR := flag.Float64("minR", -30.0, "min real")
	minI := flag.Float64("minI", -30.0, "min imag")
	maxR := flag.Float64("maxR", 30.0, "max real")
	maxI := flag.Float64("maxI", 30.0, "max imag")
	flag.Parse()

	spin := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	spin.Start()
	
	ppu := math.Pow(2, float64(*zoom))
	span := zeta.TileWidth / ppu

	x := *minR / span
	y := *minI / span

	spin.Suffix = " calculating"
	ctx := context.Background()
	algo := &zeta.Algo{}
	data := algo.Compute(ctx, complex(*minR, *minI), complex(*maxR, *maxI), zeta.TileWidth)

	t := &zeta.Tile{
		Zoom:  *zoom,
		X:     int(x),
		Y:     int(y),
		Width: zeta.TileWidth,
		Data:  data,
	}

	fname := strings.Replace(t.Filename(), ".dat.gz", ".png", -1)
	fpath := path.Join(".", fname)
	t.SavePNG(palette.DefaultPalette, fpath)
	fmt.Println("saved tile:", fpath)
}

func checkEnv() error {
	godotenv.Load()

	if os.Getenv("ZETA_NSQLOOKUP") == "" {
		return errors.New("ZETA_NSQLOOKUP is not exported")
	}

	return nil
}
