package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"zetamachine/pkg/palette"
	"zetamachine/pkg/utils"
	"zetamachine/pkg/zeta"
)

func isBackground(tile *zeta.Tile, w http.ResponseWriter) bool {
	if tile.IsBackground() {
		bkg := color.RGBA{255, 0, 0, 255} // background color of website
		rgba := image.NewRGBA(image.Rect(0, 0, zeta.TileWidth, zeta.TileWidth))
		draw.Draw(rgba, rgba.Bounds(), &image.Uniform{bkg}, image.ZP, draw.Src)

		var img image.Image = rgba
		buf := bytes.Buffer{}
		if err := png.Encode(&buf, img); err != nil {
			log.Println("Failed to encode background tile: ", err)
			return false
		}

		writePNG(w, buf.Bytes())
		return true
	}

	return false
}

func (s *Server) serveTile() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var img *image.Image

		tile, err := zeta.RequestToTile(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// don't save tiles that don't contain anything
		if isBackground(tile, w) {
			return
		}

		redo := r.URL.Query().Get("redo") != ""

		data, err := getTileData(tile, redo)
		if err != nil {
			log.Println("Failed to get tile data: ", err)
			http.Error(w, err.Error(), 500)
			return
		}

		// Render the image from the data
		img, err = tile.Render(data, palette.Original)
		if err != nil {
			log.Println("Failed to render tile data: ", err)
			http.Error(w, err.Error(), 500)
			return
		}

		// Encode the image to PNG format
		buffer := new(bytes.Buffer)
		if err := png.Encode(buffer, *img); err != nil {
			http.Error(w, "Unable to encode image: "+err.Error(), 500)
			return
		}

		writePNG(w, buffer.Bytes())
	})
}

func getTileData(tile *zeta.Tile, redo bool) (data []byte, err error) {
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

		// png, err := internalLambdaTile(tile)
		if err := farmItOut(tile); err != nil {
			return nil, err
		}

		if err := SaveData(tile); err != nil {
			log.Println("Failed to save tile data: ", err)
			return nil, err
		}
	}

	return data, nil
}

// SaveData saves the binary iteration data from a tile
func SaveData(tile *zeta.Tile) error {
	cwd, _ := os.Getwd()

	fpath := path.Join(cwd, tile.Path())
	fname := path.Join(fpath, tile.Filename())

	data, err := base64.StdEncoding.DecodeString(tile.Data)
	if err != nil {
		return err
	}

	// does not return an error if the path exists. creates the path recusively
	if err := utils.CreateFolder(fpath); err != nil {
		return err
	}

	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bytes.NewBuffer(data)
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	return nil
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

		tile := &zeta.Tile{}
		if err := json.Unmarshal(b, tile); err != nil {
			log.Println("Failed to unmarshal tile: ", err)
			http.Error(w, err.Error(), 500)
		}

		if err := renderAndEncode(tile, palette.Original, s.luts); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		b, err = json.Marshal(tile)
		if err != nil {
			s := fmt.Sprint("Failed to marshal tile: ", err)
			log.Println(s)
			http.Error(w, s, 500)
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

//
func renderAndEncode(tile *zeta.Tile, colors []color.Color, luts []*zeta.LUT) error {

	algo := &zeta.Algo{}
	data := algo.Compute(tile.Min(), tile.Max(), luts)

	img, err := tile.Render(data, colors) // returns NRGBA
	if err != nil {
		log.Println("Failed to render image: ", err)
		return errors.New(fmt.Sprint("Failed to render image: ", err))
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, *img); err != nil {
		log.Println("Unable to encode image: ", err)
		return errors.New(fmt.Sprint("Unable to encode image: ", err))
	}

	tile.Data = base64.StdEncoding.EncodeToString(data)

	return nil
}

// writeImage encodes an image 'img' in jpeg format and writes it into ResponseWriter.
func writePNG(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	if _, err := w.Write(b); err != nil {
		http.Error(w, "Unable to write image to response: "+err.Error(), 500)
		return
	}
}

// request the tile to be generated by the render farm. The tile passed in contains
// only the parameters for generating the tile, not the encoded PNG. That will
// be contained in the returned value from the remote host.
func farmItOut(tile *zeta.Tile) error {
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

func getTileBytes(url string) (*bytes.Buffer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if err := resp.Write(&buf); err != nil {
		return nil, err
	}

	return &buf, nil
}
