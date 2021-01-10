package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nsqio/go-nsq"
)

const (
	requestTopic = "request-tile"
	storeTopic   = "store-tile"
	storeChan    = "store-tile"
)

var (
	wg          *sync.WaitGroup
	sem         chan bool
	procs, zoom int
	tileCount   float64
	host, port  string
)

var producer *nsq.Producer
var consumer *nsq.Consumer

func init() {
	var err error
	log.Println("creating producer")
	// Instantiate a producer.
	config := nsq.NewConfig()
	producer, err = nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal("Could not connect to nsqd: ", err)
	}
}

func init() {
	log.Println("creating globals")
	procs = runtime.GOMAXPROCS(0)

	godotenv.Load()
	host = os.Getenv("ZETA_HOSTNAME")
	port = os.Getenv("ZETA_PORT")

	wg = &sync.WaitGroup{}
	sem = make(chan bool, procs)
}

func main() {
	checkEnv()

	zoom := flag.Int("zoom", 8, "zoom level (0-18 typically)")
	role := flag.String("role", "", "store, request, generate")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	wait := true

	switch *role {
	case "store":
		handler := &store{}
		go StartConsumer(ctx, storeTopic, "storeage", handler)
	case "request":
		request(*zoom)
		wait = false
	case "generate":
		handler := &generator{}
		go StartConsumer(ctx, requestTopic, "generator", handler)
	default:
		log.Fatal("Unknown role: ", role)
	}

	if wait {
		log.Println("Waiting for signal to exit")
		// wait for signal to exit
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Signaled to exit. Stopping NSQ")
	}

	cancel()

	if producer != nil {
		producer.Stop()
	}

	if consumer != nil {
		consumer.Stop()
	}

	log.Println("Waiting for processes to finish...")
	wg.Wait()
	log.Println("Processes complete. Stopping.")
	time.Sleep(2)
}

func checkEnv() {
	godotenv.Load()

	if os.Getenv("ZETA_NSQLOOKUP") == "" {
		log.Fatal("ZETA_NSQLOOKUP environment not set")
	}
}

func createConsumer(topic, channel string) *nsq.Consumer {
	// Instantiate a consumer that will subscribe to the provided channel.
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		log.Fatal(err)
	}

	return consumer
}
