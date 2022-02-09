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
	server, err := seed.NewCudaServer(v)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[seed] Waiting for signal to exit")

	select {
	case <-sigChan:
		log.Println("[seed] received termination request")
	case <-v.Stop():
		log.Println("[seed] process completed")
	}

	log.Println("[seed] Waiting for processes to finish...")
	v.Shutdown(10 * time.Second)
	log.Println("[seed] Processes complete.")
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
