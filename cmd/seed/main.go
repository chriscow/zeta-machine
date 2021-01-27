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
	zoom := flag.Int("zoom", 10, "maximum zoom level to generate tiles")

	maxAge := flag.Duration("max-age", time.Hour*24*30, "re-request tiles that have not completed if older")
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

func checkEnv() {
	godotenv.Load()
}

func bulkRequest(minZoom, zoom, limit int, maxAge time.Duration, s msg.Server) {
	log.Println("requesting tiles for zoom: ", zoom)

	r := s.(*Requester)

	// tileCount := int(math.Pow(2, float64(zoom+1)))
	for z := minZoom; z <= zoom; z++ {

		requested := 0
		skipped := 0

		ppu := math.Pow(2, float64(z))
		units := 1024 / ppu

		for rl := -512.0; rl < 512; rl += units {
			for im := -4096.0; im < 4096; im += units {

				patch := zeta.NewPatch(complex(rl, im), complex(rl+units, im+units))
				sent, err := r.Send(patch)
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
