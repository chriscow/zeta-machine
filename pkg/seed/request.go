package seed

import (
	"encoding/json"
	"log"
	"math"
	"os"
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
	bulbOnly bool
}

// NewRequester ...
func NewRequester(v *valve.Valve, minZoom, maxZoom int, bulbOnly bool) (*Requester, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer(os.Getenv("ZETA_NSQD"), config)
	if err != nil {
		log.Println("[requester] could not connect to nsqd: ", err)
		return nil, err
	}

	return &Requester{
		producer: p,
		valve:    v,
		minZoom:  minZoom,
		maxZoom:  maxZoom,
		bulbOnly: bulbOnly,
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

			if !r.bulbOnly {
				r.requestArms(yCount, zoom)
			}

			log.Println("zoom:", zoom, " done")
		}
		r.valve.Shutdown(0)
	}()
}

// requestBulb is a helper that requests tiles around the bulb area in the middle
// of the display near the origin. It is only valid for zoom levels of 1 or greater.
func (r *Requester) requestBulb(zoom int) int {
	log.Println("[request] -- requesting bulb --")
	xRange := math.Max(float64(zeta.TileWidth/zoom/8), 30.0)
	yRange := math.Max(float64(zeta.TileWidth/zoom/8), 20.0)

	ppu := math.Pow(2, float64(zoom))
	units := float64(zeta.TileWidth) / ppu // units per tile

	// how many patches in each direction
	xCount := int(math.Max(1, xRange/units))
	yCount := int(math.Max(1, yRange/units))

	r.requestRange(zoom, xCount, -yCount, yCount)
	return yCount
}

// requestArms is a helper to request tiles for the arms extending from the bulb
// in the imaginary axis. It is hard coded to only go to +/- 4096 on the 
// imaginary axis.
func (r *Requester) requestArms(yStart, zoom int) {
	xRange := math.Max(float64(zeta.TileWidth/zoom/8), 6.0)
	yRange := 4096.0 // same

	ppu := math.Pow(2, float64(zoom))
	units := float64(zeta.TileWidth) / ppu // units per tile

	// how many patches in each direction
	xCount := int(math.Max(2, xRange/units))
	yCount := int(math.Max(2, yRange/units))

	log.Println("[request] -- requesting positive arm -- ", yStart, yCount, " yrange, units", yRange, units)
	r.requestRange(zoom, xCount, yStart, yCount)

	log.Println("[request] -- requesting negative arm --", -yCount+1, -yStart)
	r.requestRange(zoom, xCount, -yCount+1, -yStart) // count and start need reversed
}

// requestRange generates request messages for tiles. Tiles are numbered in the
// x and y directions. The number of tiles in the x-direction (the real axis)
// are from -xCount to xCount-1
func (r *Requester) requestRange(zoom, xCount, yStart, yEnd int) (int, int) {
	log.Println("[requestRange] zoom:", zoom, " xrange ", -xCount, " to ", xCount-1)
	log.Println("\tyrange ", yStart, " to ", yEnd-1)

	sent := 0
	skipped := 0

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

			info, err := t.Exists()
			if info != nil {
				log.Println("[request] skipping. tile exists: ", t)
				skipped++
				continue
			}

			log.Println("[request] tile:", t)

			_, err = r.send(t)
			if err != nil {
				log.Fatal(err)
			}
			sent++

			select {
			case <-r.valve.Stop(): // valve is being shutdown
				return sent, skipped
			default:
			}
		}
	}

	log.Println("[requestRange] sent:", sent, " skipped:", skipped)
	return sent, skipped
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
