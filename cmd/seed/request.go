package main

import (
	"encoding/json"
	"log"
	"math"
	"zetamachine/pkg/zeta"
)

func request(zoom int) {
	log.Println("requesting tiles for zoom: ", zoom)

	tileCount := int(math.Pow(2, float64(zoom+1)) / 2)

	for y := 0; y < tileCount; y++ {
		for x := 0; x < tileCount; x++ {

			tile := zeta.Tile{Zoom: zoom, X: x, Y: y}

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
