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

var (
	zoom       int
	tileCount  float64
	host, port string
)

func main() {
	checkEnv()

	zoom := flag.Int("zoom", 4, "zoom level (0-18 typically)")
	limitY := flag.Int("y", 0, "tile limit in the imag dir. 2^zoom gets up a ways")
	limitX := flag.Int("x", 0, "tile limit in the real dir. Same as zoom get critical area")
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
		bulkRequest(*zoom, *limitY, *limitX, handler)
		wait = false
	case "generate":
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

func checkEnv() {
	godotenv.Load()
}

func bulkRequest(zoom, limitY, limitX int, s msg.Server) {
	log.Println("requesting tiles for zoom: ", zoom)

	r := s.(*msg.Requester)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("can't get working dir: ", cwd)
		return
	}

	// tileCount := int(math.Pow(2, float64(zoom+1)))
	for z := 0; z <= zoom; z++ {

		if limitY < int(math.Pow(2, float64(z))) {
			limitY = int(math.Pow(2, float64(z)))
		}

		if limitX < z*2 {
			limitX = z * 2
		}

		for y := -limitY; y < limitY; y++ {
			for x := -limitX; x < limitX; x++ {
				tile := &zeta.Tile{Zoom: z, X: x, Y: y}
				sent, err := r.Send(tile)
				if err != nil {
					log.Fatal(err)
				}

				if !sent {
					continue
				}

				log.Println("published request for: ", tile)
			}
		}
	}
}
