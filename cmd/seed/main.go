package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zetamachine/pkg/msg"

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
	zoom := flag.Int("zoom", 4, "maximum zoom level to generate tiles")

	role := flag.String("role", "", "store, request, generate")
	flag.Parse()

	if *role == "" {
		log.Fatal("Role not specified")
	}

	v := valve.New()
	wait := true

	var handler msg.Server
	var err error

	switch (*role)[:3] {
	case "sto":
		handler, err = msg.NewStore(v)
	case "req":
		handler, err = NewRequester()
		if err != nil {
			log.Fatal(err)
		}
		bulkRequest(*minZoom, *zoom, 0, *maxAge, handler)
		wait = false
	case "gen":
		handler, err = msg.NewGenerator(v)
	default:
		log.Fatal("Unknown role: ", role)
	}

	if wait {
		log.Println("[seed] Waiting for signal to exit")
		// wait for signal to exit
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("[seed] Signaled to exit. Stopping NSQ")
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
