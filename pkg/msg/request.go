package msg

import (
	"encoding/json"
	"log"
	"zetamachine/pkg/zeta"

	"github.com/nsqio/go-nsq"
)

// Requester ...
type Requester struct {
	producer *nsq.Producer
}

// NewRequester ...
func NewRequester() (*Requester, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Println("[requester] could not connect to nsqd: ", err)
		return nil, err
	}

	return &Requester{producer: p}, nil
}

// Shutdown ...
func (r *Requester) Shutdown() {
	r.producer.Stop()
}

// Send ...
func (r *Requester) Send(tile *zeta.Tile) (bool, error) {
	ok, err := tile.ShouldGenerate()
	if err != nil {
		log.Println("[requeter] could not determine if we should generate: ", tile, err)
		return ok, err
	}

	if !ok {
		log.Println("[requester] already requested tile: ", tile, "skipping.")
		return ok, nil
	}

	msg, err := json.Marshal(tile)
	if err != nil {
		log.Println("[requester] failed to marshal tile: ", tile, "\n\t", err)
		return false, err
	}

	// Synchronously publish a single message to the specified topic.
	// Messages can also be sent asynchronously and/or in batches.
	err = r.producer.Publish(RequestTopic, msg)
	if err != nil {
		log.Println("[requester] failed to publish message: ", err)
		return false, err
	}

	log.Println("[requester] tile requested: ", tile)
	return true, nil
}
