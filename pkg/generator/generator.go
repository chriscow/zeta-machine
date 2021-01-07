package generator

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
)

// Generator ...
type Generator struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stdin  io.WriteCloser
	stop   chan bool
}

// Start ...
func (g *Generator) Start() error {
	var err error
	cwd, _ := os.Getwd()
	fpath := path.Join(cwd, "zeta_machine")
	g.cmd = exec.Command(fpath, "-loop")
	g.stdout, err = g.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	g.stdin, err = g.cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := g.cmd.Start(); err != nil {
		return err
	}

	r := bufio.NewReader(g.stdout)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}

		fmt.Print(line)

		if line == "ready\n" {
			break
		}
	}

	return nil
}

// Run ...
func (g *Generator) Run(rMin, rMax, iMin, iMax, res float64) (*image.RGBA, error) {
	cmd := fmt.Sprintln(rMin, rMax, iMin, iMax, res)
	_, err := g.stdin.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 8)
	_, err = io.ReadFull(g.stdout, buf)
	if err != nil {
		return nil, err
	}

	size := binary.LittleEndian.Uint64(buf)

	r := bufio.NewReader(g.stdout)
	colors := make([]byte, size)
	_, err = io.ReadFull(r, colors)
	if err != nil {
		log.Fatal(err)
	}

	width := 60 * 50
	height := 60 * 50
	img := image.NewRGBA(image.Rect(0, 0, height, width)) // rotate 90 degrees

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := (x*width + y) * 3

			img.Set(y, x, color.NRGBA{
				R: colors[pixel+0],
				G: colors[pixel+1],
				B: colors[pixel+2],
				A: 255,
			})
		}
	}

	return img, nil
}

// Save ...
func Save(img *image.RGBA, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return err
	}

	return nil
}

// Close ...
func (g *Generator) Close() {
	g.stdin.Close()
	g.stdout.Close()
	g.cmd.Process.Kill()

	g.cmd = nil
}
