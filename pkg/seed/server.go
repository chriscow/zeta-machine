package seed

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"

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
)

type Starter interface {
	Start()
}

type CudaServer struct {
	producer *nsq.Producer
	valve    *valve.Valve
}

func NewCudaServer(v *valve.Valve) (*CudaServer, error) {

	config := nsq.NewConfig()
	p, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal("Could not connect to nsqd: ", err)
	}

	server := CudaServer{
		producer: p,
		valve:    v,
	}

	return &server, nil
}

// Start ...
func (s *CudaServer) Start() {
	go func() {
		if err := StartConsumer(s.valve.Context(), requestPatchTopic, "patch-generator", 1, s); err != nil {
			log.Fatal(err)
		}
	}()
}

// HandleMessage ...
func (s *CudaServer) HandleMessage(msg *nsq.Message) error {
	if err := s.valve.Open(); err != nil {
		log.Println("[server] failed to open valve: ", err)
		return err
	}
	defer s.valve.Close()

	var patch Patch

	if err := json.Unmarshal(msg.Body, &patch); err != nil {
		log.Println("[cuda server] Error unmarshalling msg body: ", err)
		return err
	}

	log.Println("[cuda server] Received request for tile:", patch)
	patch.Data = Generate(patch)

	json, err := json.Marshal(patch)
	if err != nil {
		log.Println("[cuda server] Error marshalling patch: ", err)
		return err
	}

	buf := &bytes.Buffer{}
	w := gzip.NewWriter(buf)
	_, err = w.Write(json)
	if err != nil {
		w.Close()
		log.Println("[cuda server] Error writing compressed json bytes to buffer:", err)
		return err
	}
	w.Close()

	log.Println("[cuda server] json size:", len(json), "after gzip:", len(buf.Bytes()))

	if err := s.producer.Publish("patch-response", buf.Bytes()); err != nil {
		log.Println("[cuda server] Error publishing response:", err)
		return err
	}

	return nil
}
