package main

import (
	web "zetamachine/pkg/web"
)

func main() {
	run()
}

func run() error {
	s := web.Server{}

	return s.Run()
}
