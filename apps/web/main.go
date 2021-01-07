package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	if err := checkEnv(); err != nil {
		log.Fatal(err)
	}

	host := os.Getenv("ZETA_HOSTNAME")
	port := os.Getenv("ZETA_PORT")

	s := server{host: host, port: port}
	// if err := s.loadLuts(); err != nil {
	// 	log.Fatal(err)
	// }

	log.Println("Listening and serving on :" + port)
	log.Println(s.run())
}

func checkEnv() error {
	godotenv.Load()

	if os.Getenv("ZETA_HOSTNAME") == "" {
		return errors.New("ZETA_HOSTNAME is not set")
	}

	if os.Getenv("ZETA_PORT") == "" {
		return errors.New("ZETA_PORT is not set")
	}

	if os.Getenv("ZETA_TILE_GENERATOR_URL") == "" {
		return errors.New("ZETA_TILE_GENERATOR_URL is not set")
	}

	if os.Getenv("ZETA_SUBDOMAINS") == "" {
		return errors.New("ZETA_SUBDOMAINS is not set")
	}

	return nil
}
func createFolder(path string) error {
	exists, err := pathExists(path)
	if err != nil {
		return err
	}

	if !exists {
		err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// exists returns whether the given file or directory exists or not
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
