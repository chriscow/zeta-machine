package main

import (
	"log"
	web "zetamachine/pkg/web"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	s := web.Server{}

	return s.Run()
}
