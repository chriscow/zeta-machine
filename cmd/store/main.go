package main

import (
	"errors"
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

	v := valve.New()
	server, err := seed.NewStore(v)
	if err != nil {
		log.Fatal(err)
	}

	server.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[store] Waiting for signal to exit")

	select {
	case <-sigChan:
		log.Println("[store] received termination request")
	case <-v.Stop():
		log.Println("[store] process completed")
	}

	log.Println("[store] Shutting down ...")
	v.Shutdown(10 * time.Second)
	log.Println("[store] Processes complete.")
}

func checkEnv() error {
	godotenv.Load()

	if os.Getenv("ZETA_NSQLOOKUP") == "" {
		return errors.New("ZETA_NSQLOOKUP is not exported")
	}

	if os.Getenv("ZETA_TILE_PATH") == "" {
		return errors.New("ZETA_TILE_PATH is not exported")
	}

	return nil
}
