package main

import (
	"flag"
	"fmt"
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

	minZoom := flag.Int("min-zoom", 0, "minimum zoom to start checking for missing tiles")
	zoom := flag.Int("zoom", 4, "maximum zoom level to generate tiles")
	limitY := flag.Int("y", 0, "tile limit in the imag dir. Defaults to +/- 2^zoom")
	limitX := flag.Int("x", 0, "tile limit in the real dir. Defaults to +/- zoom")
	maxAge := flag.Duration("max-age", time.Hour*24*30, "re-request tiles that have not completed if older")
	role := flag.String("role", "", "store, request, generate")
	flag.Parse()

	v := valve.New()
	wait := true

	var handler msg.Server
	var err error

	switch (*role)[:3] {
	case "sto":
		handler, err = msg.NewStore(v)
	case "req":
		handler, err = msg.NewRequester()
		if err != nil {
			log.Fatal(err)
		}
		bulkRequest(*minZoom, *zoom, *limitY, *limitX, *maxAge, handler)
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

func checkEnv() {
	godotenv.Load()
}

func bulkRequest(minZoom, zoom, limitY, limitX int, maxAge time.Duration, s msg.Server) {
	log.Println("requesting tiles for zoom: ", zoom)

	r := s.(*msg.Requester)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("can't get working dir: ", cwd)
		return
	}

	// tileCount := int(math.Pow(2, float64(zoom+1)))
	for z := minZoom; z <= zoom; z++ {

		requested := 0
		skipped := 0

		if limitY < int(math.Pow(2, float64(z))) {
			limitY = int(math.Pow(2, float64(z)))
		}

		if limitX < z*2 {
			limitX = z * 2
		}

		for y := -limitY; y < limitY; y++ {
			for x := -limitX; x < limitX; x++ {
				tile := &zeta.Tile{Zoom: z, X: x, Y: y}
				sent, err := r.Send(tile, maxAge)
				if err != nil {
					log.Fatal(err)
				}

				if !sent {
					skipped++
					continue
				}
				requested++
			}
		}

		fmt.Println()
		log.Println("zoom", z, "complete. sent:", requested, "skipped:", skipped)
	}
}
