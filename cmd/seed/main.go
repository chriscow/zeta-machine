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
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"zetamachine/pkg/palette"
	"zetamachine/pkg/zeta"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/briandowns/spinner"
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

	zoom := flag.Int("zoom", 0, "zoom to render")
	minR := flag.Float64("minR", -30.0, "min real")
	minI := flag.Float64("minI", -30.0, "min imag")
	maxR := flag.Float64("maxR", 30.0, "max real")
	maxI := flag.Float64("maxI", 30.0, "max imag")
	flag.Parse()

	ppu := math.Pow(2, float64(*zoom))
	span := zeta.TileWidth / ppu

	x := *minR / span
	y := *minI / span

	ctx := context.Background()
	algo := &zeta.Algo{}
	data := algo.Compute(ctx, complex(*minR, *minI), complex(*maxR, *maxI), zeta.TileWidth)

	t := &zeta.Tile{
		Zoom:  *zoom,
		X:     int(x),
		Y:     int(y),
		Width: zeta.TileWidth,
		Data:  data,
	}

	fname := strings.Replace(t.Filename(), ".dat.gz", ".png", -1)
	fpath := path.Join(".", fname)
	t.SavePNG(palette.DefaultPalette, fpath)
	fmt.Println("saved tile:", fpath)
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
	sem := make(chan bool, runtime.GOMAXPROCS(0)*2)

	spin.Suffix = " loading from " + tiles
	filepath.Walk(tiles, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}

		if info.IsDir() {
			return nil
		}

		filetype := info.Name()[len(info.Name())-3:]

		if filetype == "png" {
			return nil
		}

		// skip everything that's not .gz
		if filetype != ".gz" {
			// fmt.Println("skipping ", info.Name())
			spin.Suffix = " skipped " + info.Name()
			return nil
		}

		// we have a .gz, see if it's png exists
		pngName := strings.TrimSuffix(fpath, ".dat.gz") + ".png"
		inf, err := os.Stat(pngName)
		if inf != nil {
			// png exists
			// spin.Suffix = " png exists " + pngName
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
