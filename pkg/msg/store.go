package msg

import (
	"encoding/json"
	"log"
	"runtime"
	"zetamachine/pkg/zeta"

	"github.com/go-chi/valve"
	"github.com/nsqio/go-nsq"
)

// Store handles the storage of completed tiles to local disk
type Store struct {
	valve *valve.Valve
}

// NewStore constructs a new Store instance
func NewStore(v *valve.Valve) (*Store, error) {
	checkEnv()

	s := &Store{
		valve: v,
	}

	maxInFlight := runtime.GOMAXPROCS(0)
	go StartConsumer(v.Context(), storeTopic, storeChan, maxInFlight, s)

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

	tile := &zeta.Tile{}
	if err := json.Unmarshal(m.Body, tile); err != nil {
		log.Println("[store] failed to unmarshal tile: ", err)
		return err
	}

	if err := tile.Save(); err != nil {
		log.Println("[store] failed to save data: ", err)
		return err
	}

	markReceived(tile)

	log.Println("[store] received tile for storage: ", tile)

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}

func markReceived(tile *zeta.Tile) {
	if _, ok := queue[tile.Filename()]; ok {
		log.Println("[requester] queued tile complete: ", tile)
		delete(queue, tile.Filename())
	}
}
