package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {
	var doReorg = flag.Bool("reorg", false, "reorg tiles to new format, tar.gz the folders")
	var doGen = flag.Bool("gen", false, "generate tiles based on input")

	zoom = flag.Int("zoom", 3, "zoom level (0-18 typically)")
	xTileMin = flag.Int("tilexmin", 0, "minimum x tile")
	xTileMax = flag.Int("tilexmax", 256, "maximum x tile, inclusive")
	flag.Parse()

	if *doReorg {
		reorg()
		os.Exit(0)
	}

	if *doGen {
		wg := &sync.WaitGroup{}
		fmt.Println("Generating zoom:", *zoom, "from:", *xTileMin, "to", *xTileMax)

		wg.Add(1)
		time.AfterFunc(time.Second*3, func() {
			defer wg.Done()
			generate(*zoom, *xTileMin, *xTileMax)
		})

		wg.Wait()
	}
}
