package seed

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"zetamachine/pkg/utils"
	"zetamachine/pkg/zeta"

	"github.com/go-chi/valve"
	"github.com/nsqio/go-nsq"
)

//
// This is a simple server that takes a request to generate iteration data
// from a message received over NSQ. The iteration data is then serialized and
// published back over NSQ.
//
// The actual data generation is done via a CUDA program and linked to this one.
// (see the pkg/cuda/cuda.go file)
//
// curl -d '{"size": 1024, "min": [-30.0, -30.0], "max": [30, 30]}' 'http://127.0.0.1:4151/pub?topic=patch-request'
//
const (
	requestPatchTopic = "patch-request"
	nsqMaxMsgSize     = 1048576
)

// Starter is a basic interface that provides a Start() method
type Starter interface {
	Start()
}

// CudaServer waits for messages requesting the generation of a tile patch
// then generates the data on the GPU, splits the patch into 16 tiles
// and publishes each individual tile.
type CudaServer struct {
	producer *nsq.Producer
	valve    *valve.Valve
}

// NewCudaServer constructs a CudaServer by creating internal structures
// such as an NSQ Producer for publishing generated tiles
func NewCudaServer(v *valve.Valve) (*CudaServer, error) {

	config := nsq.NewConfig()
	p, err := nsq.NewProducer(os.Getenv("ZETA_NSQD"), config)
	if err != nil {
		log.Fatal("Could not connect to nsqd: ", err)
	}

	server := CudaServer{
		producer: p,
		valve:    v,
	}

	return &server, nil
}

// Start starts the NSQ consumer to service request messages
func (s *CudaServer) Start() {
	go func() {
		if err := utils.StartConsumer(s.valve.Context(), requestPatchTopic, "patch-generator", 1, s); err != nil {
			log.Fatal(err)
		}
	}()
}

// HandleMessage is called by the NSQ consumer when a request for a patch is received.
func (s *CudaServer) HandleMessage(msg *nsq.Message) error {
	if err := s.valve.Open(); err != nil {
		log.Println("[server] failed to open valve: ", err)
		return err
	}
	defer s.valve.Close()

	t := &zeta.Tile{}
	if err := json.Unmarshal(msg.Body, t); err != nil {
		log.Println("[cuda server] Error unmarshalling msg body: ", err)
		return err
	}

	ticker := time.NewTicker(utils.TouchSec * time.Second)
	done := make(chan bool)
	go func() {
		t.ComputeRequest(s.valve.Context())
		close(done)
	}()

loop:
	for {
		select {
		case <-done:
			break loop
		case <-ticker.C:
			log.Println("[cuda server] touching message")
			msg.Touch()
		}
	}

	// Publish the 16 tiles for storage.
	if err := s.publishTile(t); err != nil {
		// Move this patch request message to the errors topic
		if err := s.producer.Publish("patch-errors", msg.Body); err != nil {
			log.Println("[cuda server] error publishing error message:", err)
		}
	}

	return nil
}

func (s *CudaServer) publishTile(tile *zeta.Tile) error {
	json, err := json.Marshal(tile)
	if err != nil {
		log.Println("[cuda server] Error marshalling tile: ", err)
		return err
	}

	// If the compressed tile is still too large, publish the original
	// request to the `patch-errors` topic and move on
	if len(json) > nsqMaxMsgSize {
		log.Println("[cuda server] tile too large:", tile)
		return fmt.Errorf("tile to large:%d bytes", len(json)) // swallow msg from this topic
	}

	// Send the tile to be stored
	if err := s.producer.Publish("patch-response", json); err != nil {
		log.Println("[cuda server] Error publishing response:", err)
		return err
	}

	return nil
}
