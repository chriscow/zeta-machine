package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"zetamachine/pkg/cuda"
	"zetamachine/pkg/msg"
	"zetamachine/pkg/zeta"

	"github.com/joho/godotenv"
	"github.com/nsqio/go-nsq"

	"github.com/go-chi/valve"
)

//
// curl -d '{"size": 1024, "min": [-30.0, -30.0], "max": [30, 30]}' 'http://127.0.0.1:4151/pub?topic=patch-request'
//

type handler struct {
	producer *nsq.Producer
}

func main() {
	if err := checkEnv(); err != nil {
		log.Fatal(err)
	}

	v := valve.New()
	ctx := v.Context()

	config := nsq.NewConfig()
	p, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal("Could not connect to nsqd: ", err)
	}

	h := handler{
		producer: p,
	}

	go func() {
		if err := msg.StartConsumer(ctx, "patch-request", "patch-generator", 1, h); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("[cuda] Waiting for signal to exit")
	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("[cuda] Signaled to exit. Stopping NSQ ...")
	v.Shutdown(20 * time.Second)
	log.Println("[cuda] Ddone")
}

func checkEnv() error {
	godotenv.Load()

	if os.Getenv("ZETA_NSQLOOKUP") == "" {
		return errors.New("ZETA_NSQLOOKUP is not exported")
	}

	return nil
}

func (h handler) HandleMessage(msg *nsq.Message) error {

	var patch zeta.Patch

	if err := json.Unmarshal(msg.Body, &patch); err != nil {
		log.Println("[cuda] Error unmarshalling msg body: ", err)
		return err
	}

	log.Println("[cuda] Received request for tile:", patch)
	patch.Data = cuda.Generate(patch)

	json, err := json.Marshal(patch)
	if err != nil {
		log.Println("[cuda] Error marshalling patch: ", err)
		return err
	}

	buf := &bytes.Buffer{}
	w := gzip.NewWriter(buf)
	_, err = w.Write(json)
	if err != nil {
		w.Close()
		log.Println("[cuda] Error writing compressed json bytes to buffer:", err)
		return err
	}
	w.Close()

	log.Println("[cuda] json size:", len(json), "after gzip:", len(buf.Bytes()))

	if err := h.producer.Publish("patch-response", buf.Bytes()); err != nil {
		log.Println("[cuda] Error publishing response:", err)
		return err
	}

	return nil
}
