package main

import (
	"encoding/json"
	"log"
	"zetamachine/pkg/web"
	"zetamachine/pkg/zeta"

	"github.com/nsqio/go-nsq"
)

type store struct{}

func (h *store) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	tile := &zeta.Tile{}
	if err := json.Unmarshal(m.Body, tile); err != nil {
		log.Println("Failed to unmarshal tile: ", err)
		return err
	}

	if err := web.SaveData(tile); err != nil {
		log.Println("Failed to save data: ", err)
		return err
	}

	log.Println("stored tile: ", tile)

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}
