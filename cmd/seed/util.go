package main

import (
	"context"
	"os"

	"github.com/nsqio/go-nsq"
)

// StartConsumer is a helper function that starts consuming a topic from NSQ. It
// will block until the context.Done() channel closes / receives a value at which
// point it gracefully shuts down the consumer.
func StartConsumer(ctx context.Context, topic, channel string, handler nsq.Handler) error {
	// Instantiate a consumer that will subscribe to the provided channel.
	config := nsq.NewConfig()
	config.MaxInFlight = procs
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	consumer.AddHandler(handler)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = consumer.ConnectToNSQLookupd(os.Getenv("ZETA_NSQLOOKUP") + ":4161")
	// err = consumer.ConnectToNSQD("localhost:4150")
	if err != nil {
		return err
	}

	// wait for signal to exit
	<-ctx.Done()

	// Gracefully stop the consumer.
	consumer.Stop()
	return nil
}
