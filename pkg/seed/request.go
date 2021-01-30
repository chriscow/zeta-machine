package seed

import (
	"encoding/json"
	"log"
	"math"

	"github.com/go-chi/valve"

	"github.com/nsqio/go-nsq"
)

// Requester ...
type Requester struct {
	producer *nsq.Producer
	valve    *valve.Valve
	minZoom  int
	maxZoom  int
}

// NewRequester ...
func NewRequester(v *valve.Valve, minZoom, maxZoom int) (*Requester, error) {
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

		// tileCount := int(math.Pow(2, float64(zoom+1)))
		for zoom := r.minZoom; zoom <= r.maxZoom; zoom++ {

			requested := 0
			skipped := 0

			ppu := math.Pow(2, float64(zoom))
			units := float64(PatchWidth) / ppu // units per patch

		loop:
			for rl := -xRange; rl < xRange; rl += units {
				for im := -yRange; im < yRange; im += units {

					patch := NewPatch(zoom, complex(rl, im), complex(rl+units, im+units))
					sent, err := r.send(patch)
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
func (r *Requester) send(tile *Patch) (bool, error) {

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
