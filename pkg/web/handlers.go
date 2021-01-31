package web

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"zetamachine/pkg/palette"
	"zetamachine/pkg/zeta"

	"github.com/foolin/goview"
)

func (s *Server) serveIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		zoom := 4
		rl := 0.0
		im := 0.0

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

		if err := tile.Load(); err != nil {
			log.Println("Failed to get tile data: ", err)
			http.Error(w, err.Error(), 404)
			return
		}

		// Render the image from the data
		img, err = tile.Render(palette.DefaultPalette)
		if err != nil {
			log.Println("Failed to render tile data: ", err)
			http.Error(w, err.Error(), 500)
			return
		}

		// for debugging, write out the png to a file
		if err := os.MkdirAll(tile.Path(), os.ModeDir|os.ModePerm); err == nil {
			fname := strings.Replace(tile.Filename(), ".dat", ".png", -1)
			fpath := path.Join(tile.Path(), fname)

			tile.SavePNG(palette.DefaultPalette, fpath)
		}

		// send the png over the wire
		writePNG(w, img)
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
