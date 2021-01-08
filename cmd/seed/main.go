package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"zetamachine/pkg/zeta"

	"github.com/joho/godotenv"
)

var (
	wg          *sync.WaitGroup
	sem         chan bool
	procs, zoom int
	tileCount   float64
	host, port  string
)

func init() {
	procs = runtime.GOMAXPROCS(0)

	godotenv.Load()
	host = os.Getenv("ZETA_HOSTNAME")
	port = os.Getenv("ZETA_PORT")

	wg = &sync.WaitGroup{}
	sem = make(chan bool, procs)
}

func main() {
	zoom := flag.Int("zoom", 8, "zoom level (0-18 typically)")
	flag.Parse()

	tileCount = math.Pow(2, float64(*zoom+1))

	center := int(tileCount / 2)

	for y := 0; y < center; y++ {
		for x := 0; x < center; x++ {
			wg.Add(1)
			sem <- true
			go reqTile(*zoom, int(x), int(y))

			wg.Add(1)
			sem <- true
			go reqTile(*zoom, int(x), int(-y))

			wg.Add(1)
			sem <- true
			go reqTile(*zoom, int(-x), int(-y))

			wg.Add(1)
			sem <- true
			go reqTile(*zoom, int(-x), int(y))

			t := zeta.Tile{Zoom: *zoom, X: int(x), Y: int(y)}
			if math.Abs(real(t.Max())) > 5 {
				break
			}
		}
	}

	wg.Wait()
}

func reqTile(zoom, x, y int) {
	defer func() { <-sem }()
	defer wg.Done()

	if x == 0 && y == 0 {
		return
	}

	fmt.Println(zoom, x, y)

	resp, err := http.Get("http://" + host + ":" + port + "/tile/" + strconv.Itoa(zoom) + "/" + strconv.Itoa(y) + "/" + strconv.Itoa(x) + "/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatal(fmt.Sprintf("Received error code %d %s", resp.StatusCode, resp.Status))
	}
}
