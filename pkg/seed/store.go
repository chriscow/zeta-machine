package seed

import (
	"encoding/json"
	"log"
	"math"
	"runtime"
	"zetamachine/pkg/zeta"

	"github.com/go-chi/valve"
	"github.com/nsqio/go-nsq"
)

const (
	storeTopic = "patch-response"
)

// Store handles the storage of completed tiles to local disk
type Store struct {
	valve *valve.Valve
}

// NewStore constructs a new Store instance
func NewStore(v *valve.Valve) (*Store, error) {
	s := &Store{
		valve: v,
	}

	log.Println("[store] starting consumer on ", storeTopic, " `store`")
	maxInFlight := runtime.GOMAXPROCS(0)
	go StartConsumer(v.Context(), storeTopic, "store", maxInFlight, s)

	return s, nil
}

// Shutdown ...
func (s *Store) Shutdown() {
	// noop
}

// HandleMessage handles completed tiles from the Generator and stores them
// on the local disk
func (s *Store) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	if err := s.valve.Open(); err != nil {
		log.Println("[store] failed to open valve: ", err)
		return err
	}
	defer s.valve.Close()

	// units := 1024 / ppu // units per patch
	// for rl := -512.0; rl < 512; rl += units {
	// 	for im := -4096.0; im < 4096; im += units {

	// 		patch := NewPatch(complex(rl, im), complex(rl+units, im+units))

	patch := Patch{}
	if err := json.Unmarshal(m.Body, &patch); err != nil {
		log.Println("[store] failed to unmarshal patch: ", err)
		return err
	}

	log.Println("[store] received patch for storage: ", patch)

	// ppu := math.Pow(2, float64(patch.Zoom)) // pixels per unit
	// units := patch.Max[0] - patch.Min[0]    // patch spans this many units
	// upt := zeta.TileWidth / ppu             // units per tile
	// tileCount := units / ppu

	//
	// Copy the patch data into the individual tiles, two at a time
	//

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}

func saveTiles(patch Patch, rowStart int) error {
	tile1 := &zeta.Tile{
		Zoom: patch.Zoom,
		Data: make([]uint32, zeta.TileWidth*zeta.TileWidth),
		Size: zeta.TileWidth,
	}

	tile2 := &zeta.Tile{
		Zoom: patch.Zoom,
		Data: make([]uint32, zeta.TileWidth*zeta.TileWidth),
		Size: zeta.TileWidth,
	}

	// Patch data encompasses 4 tiles
	//			____ ____
	//			|   |    |
	for i := 0; i < zeta.TileWidth; i++ {
		copy(tile1.Data[rowStart+i:], patch.Data[rowStart+i:zeta.TileWidth])
		copy(tile2.Data[rowStart+i:], patch.Data[rowStart+i+zeta.TileWidth:zeta.TileWidth])
	}
	tile1.X = int(math.Floor(patch.Min[0] / float64(PatchWidth)))
	tile1.Y = int(math.Floor(patch.Min[1] / float64(PatchWidth)))
	tile2.X = tile1.X + 1
	tile2.Y = tile1.Y

	if err := tile1.Save(); err != nil {
		log.Println("[store] error saving tile:", tile1)
		log.Println("\t", err)
		return err
	}

	if err := tile2.Save(); err != nil {
		log.Println("[store] error saving tile:", tile2)
		log.Println("\t", err)
		return err
	}
}
