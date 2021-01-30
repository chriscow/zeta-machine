package web

import (
	"encoding/json"
	"log"
	"runtime"
	"zetamachine/pkg/utils"
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

	return s, nil
}

// Start ...
func (s *Store) Start() {
	log.Println("[store] starting consumer on ", storeTopic, " `store`")
	maxInFlight := runtime.GOMAXPROCS(0)
	go utils.StartConsumer(s.valve.Context(), storeTopic, "store", maxInFlight, s)
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
		log.Println("[store] failed to unmarshal patch: ", err)
		return err
	}

	if err := tile.Save(); err != nil {
		log.Println("[store] error saving tile: ", err)
		return err
	}

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}
