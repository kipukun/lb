package main

import (
	"sync"
)

type relay struct {
	Online, Primary, Disabled, Noredir bool
	Listeners, Max, Weight             int
	Status, Stream                     string
	err                                error
}

type relays struct {
	m map[string]*relay
	sync.Mutex
}

func (r *relay) activate(l int) {
	r.err = nil
	r.Online = true
	r.Listeners = l
}

func (r *relay) deactivate(e error) {
	r.Online = false
	r.Listeners = 0
	r.err = e
}

func (r *relay) status() string {
	return r.Status
}
