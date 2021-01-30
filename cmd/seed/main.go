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
	maxZoom := flag.Int("max-zoom", 4, "maximum zoom level to generate tiles")

	role := flag.String("role", "", "store, request, generate")
	flag.Parse()

	if *role == "" {
		log.Fatal("Role not specified")
	}

	var err error
	var server seed.Starter
	v := valve.New()

	switch *role {
	case "sto":
		fallthrough
	case "store":
		server, err = seed.NewStore(v)
		if err != nil {
			log.Fatal(err)
		}
	case "req":
		fallthrough
	case "request":
		server, err = seed.NewRequester(v, *minZoom, *maxZoom)
	case "gen":
		fallthrough
	case "generate":
		server, err = seed.NewCudaServer(v)
	default:
		log.Fatal("Unknown role: ", role)
	}

	if err != nil {
		log.Fatal(err)
	}
	log.Println(1)
	log.Println(server)
	server.Start()
	log.Println(2)

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

	return nil
}
