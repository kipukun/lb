package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

func main() {
	s := new(state)
	if _, err := toml.DecodeFile("config.toml", &s.c); err != nil {
		log.Fatal(err)
	}
	s.start()
}
