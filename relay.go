package main

import "sync"

type relay struct {
	Online, Primary, Disabled, Noredir bool
	Listeners, Max, Weight             int
	Status, Stream                     string
	err                                error
	sync.RWMutex
}

func (r *relay) activate(l int) {
	r.Lock()
	defer r.Unlock()
	r.err = nil
	r.Online = true
	r.Listeners = l
}

func (r *relay) deactivate(e error) {
	r.Lock()
	defer r.Lock()
	r.Online = false
	r.Listeners = 0
	r.err = e
}

func (r *relay) status() string {
	r.RLock()
	defer r.RUnlock()
	return r.Status
}
