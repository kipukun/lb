package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

func main() {
	s := new(state)
	if _, err := toml.DecodeFile("config.toml", &s.c); err != nil {
		log.Fatal(err)
	}
	if _, err := toml.DecodeFile("relays.toml", &s.relays.m); err != nil {
		log.Fatal(err)
	}
	fmt.Println("slightly")
	s.start()
}
