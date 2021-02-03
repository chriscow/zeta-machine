package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"zetamachine/pkg/palette"
	"zetamachine/pkg/seed"
	"zetamachine/pkg/zeta"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/briandowns/spinner"
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

	// getBucketRegion(context.Background(), "lattice.reticle.pasta")
	// if err := uploadToS3("D:/public/tiles/0/0/0.0.0.dat.gz", "pasta.zeta.machine/public/tiles/"); err != nil {
	// 	log.Fatal(err)
	// }
	// os.Exit(0)

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
	// 	for x := -1; x <= 1; x++ {
	// 		for y := -1; y <= 1; y++ {
	// 			t := &zeta.Tile{
	// 				Zoom:  4,
	// 				X:     x,
	// 				Y:     y,
	// 				Width: zeta.TileWidth,
	// 			}
	// 			t.ComputeRequest(context.Background())
	// 			fname := strings.Replace("patch."+t.Filename(), ".dat", ".png", -1)
	// 			fpath := path.Join(".", fname)
	// 			t.SavePNG(palette.DefaultPalette, fpath)
	// 			fmt.Println("saved tile", fpath)
	// 		}
	// 	}

	// 	os.Exit(0)
	case "decom":
		fallthrough
	case "decomp":
		fallthrough
	case "decompress":
		doDecompress()
		os.Exit(0)
	case "req":
		fallthrough
	case "request":
		server, err = seed.NewRequester(v, *minZoom, *maxZoom)
	case "gen":
		fallthrough
	case "generate":
		server, err = seed.NewCudaServer(v)
	case "sto":
		fallthrough
	case "store":
		server, err = seed.NewStore(v)
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

func decompress(r io.Reader) ([]byte, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	c, err := io.Copy(buf, zr)
	if err != nil {
		log.Println("failed to copy decompressed data: ", err, c)
		return nil, err
	}

	return buf.Bytes(), nil
}

func getBucketRegion(ctx context.Context, bucket string) error {
	sess := session.Must(session.NewSession())

	region, err := s3manager.GetBucketRegion(ctx, sess, bucket, "us-west-2")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			fmt.Fprintf(os.Stderr, "unable to find bucket %s's region not found\n", bucket)
		}
		return err
	}
	fmt.Printf("Bucket %s is in %s region\n", bucket, region)
	return nil
}

func uploadToS3(srcfname, dstfname, bucket string) error {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession())

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(srcfname)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", srcfname, err)
	}
	defer f.Close()

	// s3://pasta.zeta.machine/public/tiles/
	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstfname),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	fmt.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))
	return nil
}

func doDecompress() {
	tiles := os.Getenv("ZETA_TILE_PATH")
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.FinalMSG = "Done!"
	spin.Start()
	defer spin.Stop()

	badtiles := make([]string, 0)

	wg := &sync.WaitGroup{}
	sem := make(chan bool, runtime.GOMAXPROCS(0))

	spin.Suffix = " loading from " + tiles
	filepath.Walk(tiles, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}

		if info.IsDir() {
			return nil
		}

		filetype := info.Name()[len(info.Name())-3:]

		// delete .png
		if filetype == "png" {
			spin.Suffix = " removing " + info.Name()
			os.Remove(fpath)
			return nil
		}

		// skip everything that's not .gz
		if filetype != ".gz" {
			// fmt.Println("skipping ", info.Name())
			spin.Suffix = " skipped " + info.Name()
			return nil
		}

		spin.Suffix = " processing " + info.Name()
		wg.Add(1)
		sem <- true
		go func(info os.FileInfo) {
			tile, err := zeta.TileFromFilename(info.Name())
			if err != nil {
				log.Fatal("failed to load tile: ", info.Name(), "\n\t", err)
			}

			pngName := strings.TrimSuffix(info.Name(), ".dat.gz")
			tile.SavePNG(palette.DefaultPalette, path.Join(tile.Path(), pngName+".png"))
			wg.Done()
			<-sem
		}(info)

		return nil
	})

	spin.Suffix = " waiting for saves"

	if len(badtiles) > 0 {
		fmt.Println("Bad Tiles:")
		for i := range badtiles {
			fmt.Println(badtiles[i])
		}
	}
}
