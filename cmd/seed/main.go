package main

import (
	"flag"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zetamachine/pkg/msg"
	"zetamachine/pkg/zeta"

	"github.com/go-chi/valve"
	"github.com/joho/godotenv"
)

const (
	requestTopic = "request-tile"
	storeTopic   = "store-tile"
	storeChan    = "store-tile"
)

var (
	zoom       int
	tileCount  float64
	host, port string
)

func main() {
	checkEnv()

	zoom := flag.Int("zoom", 8, "zoom level (0-18 typically)")
	role := flag.String("role", "", "store, request, generate")
	flag.Parse()

	v := valve.New()
	wait := true

	var handler msg.Server
	var err error

	switch *role {
	case "store":
		handler, err = msg.NewStore(v)
	case "request":
		handler, err = msg.NewRequester()
		if err != nil {
			log.Fatal(err)
		}
		bulkRequest(*zoom, handler)
		wait = false
	case "generate":
		handler, err = msg.NewGenerator(v)
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

	log.Println("Waiting for processes to finish...")
	v.Shutdown(20 * time.Second)
	log.Println("Processes complete. Stopping.")
}

func checkEnv() {
	godotenv.Load()
}

func bulkRequest(zoom int, s msg.Server) {
	log.Println("requesting tiles for zoom: ", zoom)

	r := s.(*msg.Requester)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("can't get working dir: ", cwd)
	}

	tileCount := int(math.Pow(2, float64(zoom+1)))
	limit := tileCount / (zoom + 2)

	for y := -limit; y < limit; y++ {
		for x := -limit; x < limit; x++ {

			tile := &zeta.Tile{Zoom: zoom, X: x, Y: y}
			if err := r.Send(tile); err != nil {
				log.Fatal(err)
			}

			log.Println("published request for: ", tile)
		}
	}
}
