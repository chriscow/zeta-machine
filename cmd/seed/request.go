package main

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"path"
	"zetamachine/pkg/utils"
	"zetamachine/pkg/zeta"
)

func request(zoom int) {
	log.Println("requesting tiles for zoom: ", zoom)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("can't get working dir: ", cwd)
	}

	tileCount := int(math.Pow(2, float64(zoom+1)))
	limit := tileCount / (zoom + 2)

	for y := -limit; y < limit; y++ {
		for x := -limit; x < limit; x++ {

			tile := zeta.Tile{Zoom: zoom, X: x, Y: y}

			// see if we already have this tile
			fname := path.Join(cwd, tile.Path(), tile.Filename())
			exists, err := utils.PathExists(fname)
			if err != nil {
				log.Fatal("error checking for path: ", err)
			}

			if exists {
				continue
			}

			msg, err := json.Marshal(tile)
			if err != nil {
				log.Fatal("Failed to marshal tile: ", tile, "\n\t", err)
			}

			// Synchronously publish a single message to the specified topic.
			// Messages can also be sent asynchronously and/or in batches.
			err = producer.Publish(requestTopic, msg)
			if err != nil {
				log.Fatal("Failed to publish message: ", err)
			}

			log.Println("published request for: ", tile)
		}
	}
}
