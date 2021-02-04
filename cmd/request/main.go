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

	minZoom := flag.Int("min-zoom", 0, "minimum zoom to start checking for missing tiles")
	maxZoom := flag.Int("max-zoom", 0, "maximum zoom level to generate tiles")

	flag.Parse()

	var err error
	v := valve.New()

	server, err := seed.NewRequester(v, *minZoom, *maxZoom)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[request] Waiting for signal to exit")

	select {
	case <-sigChan:
		log.Println("[request] received termination request")
	case <-v.Stop():
		log.Println("[request] process completed")
	}

	log.Println("[request] Waiting for processes to finish...")
	v.Shutdown(10 * time.Second)
	log.Println("[request] Processes complete.")
}

func checkEnv() error {
	godotenv.Load()

	if os.Getenv("ZETA_NSQLOOKUP") == "" {
		return errors.New("ZETA_NSQLOOKUP is not exported")
	}

	return nil
}
