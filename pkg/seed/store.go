package seed

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"runtime"
	"strings"
	"time"
	"zetamachine/pkg/palette"
	"zetamachine/pkg/utils"
	"zetamachine/pkg/zeta"

	"github.com/briandowns/spinner"
	"github.com/go-chi/valve"
	"github.com/nsqio/go-nsq"
)

// aws s3 sync ./9 s3://pasta.zeta.machine/public/tiles/9 --dryrun --size-only

const (
	storeTopic = "patch-response"
)

// Store handles the storage of completed tiles to local disk
type Store struct {
	valve *valve.Valve
	spin  *spinner.Spinner
}

// NewStore constructs a new Store instance
func NewStore(v *valve.Valve) (*Store, error) {
	s := &Store{
		valve: v,
		spin:  spinner.New(spinner.CharSets[43], 100*time.Millisecond),
	}

	return s, nil
}

// Start ...
func (s *Store) Start() {
	log.Println("[store] starting consumer on ", storeTopic, " `store`")
	maxInFlight := runtime.GOMAXPROCS(0) * 2
	go utils.StartConsumer(s.valve.Context(), storeTopic, "store", maxInFlight, s)
	s.spin.Start()
	s.spin.Suffix = fmt.Sprintf(" saving tiles maxInFlight: %d", maxInFlight)
}

func (s *Store) Close() error {
	s.spin.Stop()
	return nil
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

	// s.spin.Suffix = " saving " + tile.Filename()
	if err := tile.Save(); err != nil {
		log.Println("[store] error saving tile: ", err)
		return err
	}

	fname := strings.TrimSuffix(tile.Filename(), ".dat.gz")
	fpath := path.Join(tile.Path(), fname+".png")
	s.spin.Suffix = " saving png " + fpath
	if err := tile.SavePNG(palette.DefaultPalette, fpath); err != nil {
		log.Println("[store] error saving tile: ", err)
	}

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	// s.spin.Suffix = " waiting for tile"
	return nil
}
