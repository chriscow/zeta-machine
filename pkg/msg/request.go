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
func (r *Requester) Send(tile *zeta.Tile) error {
	exists, err := tile.Exists()
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	qm.Lock()
	defer qm.Unlock()

	if _, ok := queue[tile.Filename()]; ok {
		log.Println("[requester] already requested tile: ", tile, "skipping.")
		return nil
	}

	queue[tile.Filename()] = true

	msg, err := json.Marshal(tile)
	if err != nil {
		log.Println("[requester] failed to marshal tile: ", tile, "\n\t", err)
		return err
	}

	// Synchronously publish a single message to the specified topic.
	// Messages can also be sent asynchronously and/or in batches.
	err = r.producer.Publish(requestTopic, msg)
	if err != nil {
		log.Println("[requester] failed to publish message: ", err)
		return err
	}

	log.Println("[requester] tile requested: ", tile)
	return nil
}
