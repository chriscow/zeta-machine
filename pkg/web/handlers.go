package web

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"zetamachine/pkg/palette"
	"zetamachine/pkg/zeta"

	"github.com/foolin/goview"
)

var (
	// ErrTileQueued is returned if the tile was asynchonously queued for rendering
	ErrTileQueued = errors.New("Tile queued for rendering")
)

// isBlank determines if the tile should be rendered as a blank, solid color tile.
// Returns true if the response was written and no further action is needed.
func isBlank(tile *zeta.Tile, w http.ResponseWriter) bool {
	if tile.IsBackground() || tile.Zoom == -1 {
		bkg := color.RGBA{0x00, 0x3c, 0xff, 0xff} // background color of website 003cff

		img := tile.RenderSolid(bkg)

		writePNG(w, img)
		return true
	}

	return false
}

func (s *Server) serveIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var zoom int
		var rl, im float64
		var err error

		zoom, err = strconv.Atoi(os.Getenv("ZETA_DEFAULT_ZOOM"))
		if err != nil {
			http.Error(w, "ZETA_DEFAULT_ZOOM invalid environment: "+err.Error(), 500)
			return
		}

		rl, err = strconv.ParseFloat(os.Getenv("ZETA_DEFAULT_REAL"), 64)
		if err != nil {
			http.Error(w, "ZETA_DEFAULT_REAL invalid environment: "+err.Error(), 500)
			return
		}

		im, err = strconv.ParseFloat(os.Getenv("ZETA_DEFAULT_IMAG"), 64)
		if err != nil {
			http.Error(w, "ZETA_DEFAULT_IMAG invalid environment: "+err.Error(), 500)
			return
		}

		if r.URL.Query().Get("zoom") != "" {
			zoom, err = strconv.Atoi(r.URL.Query().Get("zoom"))
		}
		if r.URL.Query().Get("real") != "" {
			rl, err = strconv.ParseFloat(r.URL.Query().Get("real"), 64)
		}
		if r.URL.Query().Get("imag") != "" {
			im, err = strconv.ParseFloat(r.URL.Query().Get("imag"), 64)
		}

		goview.DefaultConfig.DisableCache = true
		err = goview.Render(w, http.StatusOK, "index.html", goview.M{
			"host":       s.host + ":" + s.port,
			"subdomains": strings.Join(s.subdomains, ""),
			"zoom":       zoom,
			"real":       rl,
			"imag":       im,
			"tileSize":   zeta.TileWidth,
		})

		if err != nil {
			fmt.Fprintf(w, "Render index error: %v!", err)
			http.Error(w, "DEFAULT_IMAG invalid environment: "+err.Error(), 500)
		}
	}
}

func (s *Server) serveTile() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var img image.Image

		tile, err := zeta.RequestToTile(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// don't save tiles that don't contain anything
		if isBlank(tile, w) {
			return
		}

		redo := r.URL.Query().Get("redo") != ""

		data, err := s.getTileData(tile, redo)
		if err != nil {
			if err == ErrTileQueued {
				img = tile.RenderSolid(palette.BackgroundColor)
			} else {
				log.Println("Failed to get tile data: ", err)
				http.Error(w, err.Error(), 500)
				return
			}
		}

		// If an image wasn't rendered above, and we have some data ...
		if img == nil && data != nil {
			// Render the image from the data
			img, err = tile.Render(data, palette.DefaultPalette)
			if err != nil {
				log.Println("Failed to render tile data: ", err)
				http.Error(w, err.Error(), 500)
				return
			}

			saveTmpPNG(tile, img)
		}

		writePNG(w, img)
	})
}

func saveTmpPNG(tile *zeta.Tile, img image.Image) {
	buf := &bytes.Buffer{}
	err := png.Encode(buf, img)
	if err != nil {
		log.Println("[saveTmpPNG] failed to encode: ", err)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Println("[saveTmpPNG] failed to get wd: ", err)
		return
	}

	fname := strings.Replace(tile.Filename(), ".dat", ".tmp.png", -1)
	fpath := path.Join(cwd, tile.Path(), fname)
	info, _ := os.Stat(fpath)
	if info == nil {
		return
	}

	f, err := os.Create(fpath)
	if err != nil {
		log.Println("[saveTmpPNG] failed to open: ", fpath, err)
		return
	}
	defer f.Close()

	i, err := io.Copy(f, buf)
	if err != nil {
		log.Println("[saveTmpPNG] failed to copy: ", err, i)
		return
	}
}

func (s *Server) getTileData(tile *zeta.Tile, redo bool) (data []byte, err error) {
	cwd, _ := os.Getwd()

	fpath := path.Join(cwd, tile.Path())
	fname := path.Join(fpath, tile.Filename())

	if _, err := os.Stat(fname); err == nil && !redo {
		f, err := os.Open(fname)
		if err != nil {
			log.Println("Failed to open data file: ", err)
			return nil, err
		}
		defer f.Close()

		data, err = ioutil.ReadAll(f)
		if err != nil {
			log.Println("Failed to read data file: ", err)
			return nil, err
		}

	} else {

		// The plan is to only generate tiles for authorized users
		// TODO: Do an auth check and queue up tile rendering. Return a temp tile

		// Tile data file does not exist.

		// hires, _ := tile.Downsample()
		// if hires != nil {
		// 	tile.Data = hires.Data
		// } else {
		// 	lowres, _ := tile.Upsample()
		// }

		// if os.Getenv("ZETA_GENERATE_VIA") == "nsq" {
		// 	_, err := s.requester.Send(tile)
		// 	if err != nil {
		// 		log.Println("[server] Failed to send NSQ generate request: ", err)
		// 		return nil, err
		// 	}

		// 	return nil, ErrTileQueued

		// } else {
		// reqGenHTTP alters the passed in tile with the resulting data
		if err := reqGenHTTP(tile); err != nil {
			log.Println("[server] Failed to send HTTP generate request: ", err)
			return nil, err
		}

		// Save and decode the data
		if err := tile.Save(); err != nil {
			log.Println("Failed to save tile data: ", err)
			return nil, err
		}

		data, err = base64.StdEncoding.DecodeString(tile.Data)
		if err != nil {
			log.Println("Failed to decode tile data: ", err)
			return nil, err
		}
		// }
	}

	return data, nil
}

// A generate request comes as a posted Tile JSON, without the encoded image of course.
// generateTile() uses the posted tile as arguments for generating the image
func (s *Server) generateTile() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// the body contains a JSON formatted Tile that we will use as
		// parameters for generating the actual image
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Failed to read body: ", err)
			http.Error(w, err.Error(), 500)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		if err != nil {
			log.Println("Failed to create timeout context: ", err)
			http.Error(w, err.Error(), 500)
		}
		defer cancel()

		b, err = zeta.ComputeRequest(ctx, b, s.luts)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		if _, err := w.Write(b); err != nil {
			http.Error(w, "Unable to write image to response: "+err.Error(), 500)
			return
		}
	})
}

// writeImage encodes an image 'img' in jpeg format and writes it into ResponseWriter.
func writePNG(w http.ResponseWriter, img image.Image) {
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		http.Error(w, "Unable to encode image: "+err.Error(), 500)
		return
	}

	b := buffer.Bytes()

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))

	if _, err := w.Write(b); err != nil {
		http.Error(w, "Unable to write image to response: "+err.Error(), 500)
		return
	}
}

// request the tile via HTTP to be generated by the render farm.
// The tile passed in contains only the parameters for generating the tile, not
// the encoded data. The data will be contained in the returned value from
// the remote host.
func reqGenHTTP(tile *zeta.Tile) error {
	url := os.Getenv("ZETA_TILE_GENERATOR_URL")

	buf, err := json.Marshal(tile)
	if err != nil {
		return err
	}

	r := bytes.NewReader(buf)
	resp, err := http.Post(url, "application/json", r)
	if err != nil {
		log.Println("Request for tile generation failed: ", err)
		return err
	}

	if resp.StatusCode != 200 {
		log.Println("Failed request for tile generation: ", resp.StatusCode, resp.Status)
		return errors.New(fmt.Sprint("Failed request for tile generation: ", tile, resp.StatusCode, resp.Status))
	}

	// the response body contains a JSON encoded tile with a base64 encoded image
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read all bytes from response: ", err)
		return err
	}

	if err := json.Unmarshal(b, tile); err != nil {
		log.Println("Failed to unmarshal tile: ", err)
		return err
	}

	return nil
}
