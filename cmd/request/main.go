package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"zetamachine/pkg/seed"

	"github.com/briandowns/spinner"
	"github.com/go-chi/valve"
	"github.com/joho/godotenv"
)

var (
	zoom       int
	tileCount  float64
	host, port string
)

func main() {
	if err := checkEnv(); err != nil {
		log.Fatal(err)
	}

	minZoom := flag.Int("min-zoom", 1, "minimum zoom to start checking for missing tiles")
	maxZoom := flag.Int("max-zoom", 1, "maximum zoom level to generate tiles")
	bulbOnly := flag.Bool("bulb-only", true, "only generate the bulb")
	flag.Parse()

	log.Println("Arguments  min-zoom:", *minZoom, "max-zoom:", *maxZoom, "bulb only: ", *bulbOnly)

	if *minZoom <= 0 {
		log.Fatal("min-zoom must be greater than zero")
	}

	if *maxZoom <= 0 {
		log.Fatal("max-zoom must be greater than zero")
	}

	var err error
	v := valve.New()
	spin := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	spin.Start()

	server, err := seed.NewRequester(v, *minZoom, *maxZoom, *bulbOnly)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("[request] received termination request")
	case <-v.Stop():
		log.Println("[request] process completed")
	}

	spin.Stop()
	v.Shutdown(10 * time.Second)
}

func checkEnv() error {
	godotenv.Load()

	if os.Getenv("ZETA_NSQLOOKUP") == "" {
		return errors.New("ZETA_NSQLOOKUP is not exported")
	}

	if os.Getenv("ZETA_NSQD") == "" {
		return errors.New("ZETA_NSQD is not exported")
	}
	return nil
}
