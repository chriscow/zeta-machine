package seed

import (
	"encoding/json"
	"log"
	"math"
	"zetamachine/pkg/zeta"

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

		log.Println("[request] zoom:", r.minZoom, "-", r.maxZoom)

		// tileCount := int(math.Pow(2, float64(zoom+1)))
		for zoom := r.minZoom; zoom <= r.maxZoom; zoom++ {

			yCount := r.requestBulb(zoom)
			r.requestArms(yCount, zoom)

			log.Println("zoom:", zoom, " done")
		}
		r.valve.Shutdown(0)
	}()
}

func (r *Requester) requestBulb(zoom int) int {
	log.Println("[request] -- requesting bulb --")
	xRange := 30.0
	yRange := 20.0

	ppu := math.Pow(2, float64(zoom))
	units := float64(zeta.TileWidth) / ppu // units per tile

	// handles the case for zoom == 0 because
	// xRange is less than a single tile, set to whole tile
	xrange := xRange
	yrange := yRange
	if zoom < 4 {
		xrange = math.Max(zeta.TileWidth, xRange)
		yrange = math.Max(zeta.TileWidth, yRange)
	}

	// how many patches in each direction
	xCount := int(xrange / units)
	yCount := int(yrange / units)

	r.requestRange(zoom, xCount, -yCount, yCount)
	return yCount
}

func (r *Requester) requestArms(yStart, zoom int) {
	xRange := 16.0   // in each direction: -512 -> 512
	yRange := 4096.0 // same

	ppu := math.Pow(2, float64(zoom))
	units := float64(zeta.TileWidth) / ppu // units per tile

	// handles the case for zoom == 0 because
	// xRange is less than a single tile, set to whole tile
	xrange := xRange
	yrange := yRange
	if zoom < 4 {
		xrange = math.Max(zeta.TileWidth, xRange)
		yrange = math.Max(zeta.TileWidth, yRange)
	}

	// how many patches in each direction
	xCount := int(xrange / units)
	yCount := int(yrange / units)

	log.Println("[request] -- requesting positive arm --")
	r.requestRange(zoom, xCount, yStart, yCount)

	log.Println("[request] -- requesting negative arm --")
	r.requestRange(zoom, xCount, -yCount+1, -yStart) // count and start need reversed
}

func (r *Requester) requestRange(zoom, xCount, yStart, yEnd int) {
	log.Println("[request] xrange ", -xCount, " to ", xCount-1)
	log.Println("[request] yrange ", yStart, " to ", yEnd-1)

	for x := -xCount; x < xCount; x++ {
		for y := yStart; y < yEnd; y++ {

			// rl := -xr + units*float64(x+xCount)
			// im := -yRange + units*float64(y+yCount)

			t := &zeta.Tile{
				Zoom:  int(zoom),
				X:     x,
				Y:     y,
				Width: zeta.TileWidth,
			}

			log.Println("[request] tile:", t)

			_, err := r.send(t)
			if err != nil {
				log.Fatal(err)
			}

			select {
			case <-r.valve.Stop(): // valve is being shutdown
				return
			default:
			}
		}
	}
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
