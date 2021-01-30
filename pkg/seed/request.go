package seed

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/nsqio/go-nsq"
)

// Requester ...
type Requester struct {
	producer *nsq.Producer
}

// NewRequester ...
func NewRequester() (*Requester, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Println("[requester] could not connect to nsqd: ", err)
		return nil, err
	}

	return &Requester{producer: p}, nil
}

// Close ...
func (r *Requester) Close() {
	r.producer.Stop()
}

// Send ...
func (r *Requester) Send(tile *Patch) (bool, error) {

	msg, err := json.Marshal(tile)
	if err != nil {
		log.Println("[requester] failed to marshal tile: ", tile, "\n\t", err)
		return false, err
	}

	// Synchronously publish a single message to the specified topic.
	// Messages can also be sent asynchronously and/or in batches.
	err = r.producer.Publish(requestPatchTopic, msg)
	if err != nil {
		log.Println("[requester] failed to publish message: ", err)
		return false, err
	}

	return true, nil
}

func (r *Requester) bulkRequest(minZoom, zoom, limit int, maxAge time.Duration) {
	log.Println("requesting tiles for zoom: ", zoom)

	// tileCount := int(math.Pow(2, float64(zoom+1)))
	for z := minZoom; z <= zoom; z++ {

		requested := 0
		skipped := 0

		ppu := math.Pow(2, float64(z))
		units := float64(PatchWidth) / ppu // units per patch

		for rl := -xRange; rl < xRange; rl += units {
			for im := -yRange; im < yRange; im += units {

				patch := NewPatch(z, complex(rl, im), complex(rl+units, im+units))
				sent, err := r.Send(patch)
				if err != nil {
					log.Fatal(err)
				}

				if !sent {
					skipped++
					continue
				}
				requested++
			}
		}

		fmt.Println()
		log.Println("zoom", z, "complete. sent:", requested, "skipped:", skipped)
	}
}
