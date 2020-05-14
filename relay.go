package main

type relay struct {
	Online, Primary, Disabled, Noredir bool
	Listeners, Max, Weight             int
	Status, Stream                     string
	err                                error
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
