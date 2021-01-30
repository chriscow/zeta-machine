package seed

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"time"

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

type cudaServer struct {
	producer *nsq.Producer
	v        *valve.Valve
}

func NewCudaServer() (*cudaServer, error) {
	v := valve.New()

	config := nsq.NewConfig()
	p, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal("Could not connect to nsqd: ", err)
	}

	server := cudaServer{
		producer: p,
		v:        v,
	}

	go func() {
		if err := StartConsumer(v.Context(), requestPatchTopic, "patch-generator", 1, &server); err != nil {
			log.Fatal(err)
		}
	}()

	return &server, nil
}

func (s *cudaServer) Close() {
	s.v.Shutdown(20 * time.Second)
}

func (s *cudaServer) HandleMessage(msg *nsq.Message) error {
	if err := s.v.Open(); err != nil {
		log.Println("[server] failed to open valve: ", err)
		return err
	}
	defer s.v.Close()

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
