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

	role := flag.String("role", "", "request, generate")
	flag.Parse()

	if *role == "" {
		log.Fatal("Role not specified")
	}

	var err error
	var server seed.Starter
	v := valve.New()

	switch *role {
	// case "make":
	// 	p := seed.NewPatch(0, 0, complex(-512, -512), complex(512, 512), 0, 0, seed.PatchWidth)
	// 	p.Generate(context.Background())
	// 	t := p.ToTile()
	// 	fname := strings.Replace("patch."+t.Filename(), ".dat", ".png", -1)
	// 	fpath := path.Join(".", fname)
	// 	t.SavePNG(palette.DefaultPalette, fpath)
	// 	fmt.Println("saved patch", fpath)

	// 	tiles, err := p.Split()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	for i := range tiles {
	// 		fname := strings.Replace(tiles[i].Filename(), ".dat", ".png", -1)
	// 		fpath := path.Join(".", fname)
	// 		fmt.Println(fpath)
	// 		tiles[i].SavePNG(palette.DefaultPalette, fpath)
	// 	}
	// 	os.Exit(0)

	case "req":
		fallthrough
	case "request":
		server, err = seed.NewRequester(v, uint8(*minZoom), uint8(*maxZoom))
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

	return nil
}
