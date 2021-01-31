package seed

import (
	"encoding/json"
	"log"
	"math"
	"zetamachine/pkg/zeta"

	"github.com/go-chi/valve"
	"github.com/nsqio/go-nsq"
)

const (
	// These are the ranges we want to render
	xRange = 512.0  // in each direction: -512 -> 512
	yRange = 4096.0 // same
)

// Requester ...
type Requester struct {
	producer *nsq.Producer
	valve    *valve.Valve
	minZoom  uint8
	maxZoom  uint8
}

// NewRequester ...
func NewRequester(v *valve.Valve, minZoom, maxZoom uint8) (*Requester, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Println("[requester] could not connect to nsqd: ", err)
		return nil, err
	}

	return &Requester{
		producer: p,
		valve:    v,
		minZoom:  minZoom,
		maxZoom:  maxZoom,
	}, nil
}

// Start ...
func (r *Requester) Start() {
	go func() {
		defer r.producer.Stop()

		log.Println("[request] zoom:", r.minZoom, "-", r.maxZoom)

		var count uint64

		// tileCount := int(math.Pow(2, float64(zoom+1)))
		for zoom := r.minZoom; zoom <= r.maxZoom; zoom++ {

			requested := 0
			skipped := 0

			ppu := math.Pow(2, float64(zoom))
			units := float64(zeta.TileWidth) / ppu // units per tile

			// handles the case for zoom == 0 because
			// xRange is less than a single tile, set to whole tile
			xRange := math.Max(zeta.TileWidth, xRange)

			// how many patches in each direction
			xCount := int(xRange / units)
			yCount := int(yRange / units)

		loop:
			for x := -xCount; x < xCount; x++ {
				for y := -yCount; y < yCount; y++ {

					// rl := -xr + units*float64(x+xCount)
					// im := -yRange + units*float64(y+yCount)

					t := &zeta.Tile{
						Zoom:  int(zoom),
						X:     x,
						Y:     y,
						Width: zeta.TileWidth,
					}

					// (count, zoom, complex(rl, im), complex(rl+units, im+units), x, y, width)
					count++

					log.Println("[request] tile:", t)

					sent, err := r.send(t)
					if err != nil {
						log.Fatal(err)
					}

					if !sent {
						skipped++
						continue
					}
					requested++

					select {
					case <-r.valve.Stop(): // valve is being shutdown
						break loop
					default:
					}
				}
			}
			log.Println("zoom:", zoom, "done. sent:", requested, "skipped:", skipped)
		}
		r.valve.Shutdown(0)
	}()
}

// Send ...
func (r *Requester) send(tile *zeta.Tile) (bool, error) {

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
