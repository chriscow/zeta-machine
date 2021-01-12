package msg

import (
	"log"
	"runtime"
	"time"
	"zetamachine/pkg/zeta"

	"github.com/go-chi/valve"

	"github.com/nsqio/go-nsq"
)

// Generator ...
type Generator struct {
	valve    *valve.Valve
	producer *nsq.Producer
}

// NewGenerator ...
func NewGenerator(v *valve.Valve) (*Generator, error) {
	checkEnv()

	var err error
	config := nsq.NewConfig()
	p, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal("Could not connect to nsqd: ", err)
	}

	g := &Generator{valve: v, producer: p}

	maxInFlight := 1 // handle this many messages in parallel
	if runtime.GOMAXPROCS(0) > 16 {
		maxInFlight = runtime.GOMAXPROCS(0)
	}

	go StartConsumer(v.Context(), requestTopic, genChan, maxInFlight, g)

	return g, nil
}

// Shutdown cleanly shuts down
func (g *Generator) Shutdown() {
	g.producer.Stop()
}

// HandleMessage implements the nsq.Handler interface.
func (g *Generator) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	if err := g.valve.Open(); err != nil {
		log.Println("[generator] error opening valve while handling message: ", err)
		return err
	}
	defer g.valve.Close()

	var res []byte

	start := time.Now()

	done := make(chan bool)
	ticker := time.NewTicker(time.Second * touchSec)

	go func() {
		defer func() { done <- true }()
		var err error
		res, err = zeta.ComputeRequest(m.Body, nil)
		if err != nil {
			log.Println("[generate] ComputeRequest failed: ", err)
			return
		}
	}()

	for {
		select {
		case <-done:
			if err := g.producer.Publish(storeTopic, res); err != nil {
				log.Println("[generator] failed to publish results: ", err)
				return err
			}

			log.Println("[generator] completed in", time.Since(start))
			return nil

		case <-ticker.C:
			m.Touch()
		}
	}
}
