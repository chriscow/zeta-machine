package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"time"
	"zetamachine/pkg/zeta"

	"github.com/nsqio/go-nsq"
)

const touchSec = 30 // touch the message every so often

type generator struct{}

// HandleMessage implements the Handler interface.
func (h *generator) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	sem <- true
	wg.Add(1)

	tile := &zeta.Tile{}
	if err := json.Unmarshal(m.Body, tile); err != nil {
		log.Println("Failed to unmarshal tile: ", err)
		return err
	}

	go func(tile *zeta.Tile) {
		log.Println("Generate requested: ", tile)
		start := time.Now()

		algo := &zeta.Algo{}

		var data []byte
		log.Println("Starting compute")
		data = algo.Compute(tile.Min(), tile.Max(), nil)
		tile.Data = base64.StdEncoding.EncodeToString(data)

		b, err := json.Marshal(tile)
		if err != nil {
			log.Println("Failed to marshal tile: ", err)
			return
		}

		if err := producer.Publish(storeTopic, b); err != nil {
			log.Println("Failed to publish results: ", err)
			return
		}

		log.Println("Generate completed: ", tile, "in", time.Since(start))
		<-sem
		wg.Done()
	}(tile)

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}
