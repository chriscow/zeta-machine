package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func Tar(source, target, zoom string) error {
	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s-%s.tar", zoom, filename))
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		// fix up baseDir to remove the tile zoom folder
		// and also remove the CWD prefix
		baseDir = source
		// baseDir = path.Dir(baseDir)

		cwd, _ := os.Getwd()
		baseDir = strings.Replace(baseDir, cwd, "", -1)
		baseDir = strings.TrimPrefix(baseDir, "/")
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

func Gzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}

	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s.gz", filename))
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filename
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)
	return err
}

func reorg() {
	wg := &sync.WaitGroup{}

	for zoom := 0; zoom <= 18; zoom++ {
		packZoom(zoom, wg)
	}

	wg.Wait()
	fmt.Println("done")
}

func packZoom(zoom int, wg *sync.WaitGroup) {
	var t []string
	lastx := ""

	cwd, _ := os.Getwd()
	fpath := path.Join(cwd, "public/tiles")

	tmpDir := path.Join(cwd, "tmp")
	if err := createFolder(tmpDir); err != nil {
		log.Fatal(tmpDir, " ", err)
	}

	dir := path.Join(fpath, strconv.Itoa(zoom))
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	log.Println("Moving zoom", zoom)

	for _, f := range files {
		if !strings.Contains(f.Name(), ".png") {
			continue
		}
		t = strings.Split(f.Name(), "-")
		// Make a new directory if we are on a new zoom
		if t[1] != lastx {
			if lastx != "" {
				// wg.Add(1)
				// packXFolder(cwd, path.Join(dir, lastx), tmpDir, lastx, zoom, wg)
			}

			os.Mkdir(path.Join(dir, t[1]), os.ModeDir|os.ModePerm)
			log.Println("\nMoving zoom", zoom, "x:", t[1])
			lastx = t[1]
		}

		from := path.Join(dir, f.Name())
		to := path.Join(dir, t[1], f.Name())
		os.Rename(from, to)
		fmt.Print(".")
	}

	// take care of the final folder
	// wg.Add(1)
	// packXFolder(cwd, path.Join(dir, t[1]), tmpDir, t[1], zoom, wg)

	fmt.Println("")
}

func packXFolder(cwd, source, dest, lastx string, zoom int, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := Tar(source, dest, strconv.Itoa(zoom)); err != nil {
		log.Fatal("tar dir", source, err)
	}

	if err := Gzip(path.Join(dest, strconv.Itoa(zoom)+"-"+lastx+".tar"), dest); err != nil {
		log.Fatal("gzip", path.Join(cwd, strconv.Itoa(zoom)+"."+lastx+".tar.gz"), err)
	}

	os.RemoveAll(path.Join(cwd, "public/tiles", lastx))
	os.Remove(path.Join(dest, strconv.Itoa(zoom)+"-"+lastx+".tar"))
}
