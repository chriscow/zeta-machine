package main

import (
	"zetamachine/pkg/web"
)

func main() {
	run()
}

func run() error {
	s := web.Server{}

	return s.Run()
}
