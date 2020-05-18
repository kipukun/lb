package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

type config struct {
	Addr, Fallback string
}

type state struct {
	serv      *http.Server
	c         config
	listeners int
	min       float64
	current   atomic.Value
	relays    relays
}

func health(c *http.Client, relay *relay, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := c.Get(relay.status())
	if err != nil {
		relay.deactivate(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		relay.deactivate(err)
		return
	}
	l, err := parsexml(body)
	if err != nil {
		relay.deactivate(err)
		return
	}
	relay.activate(l)
	return
}

func (s *state) choose() {
	s.relays.Lock()
	defer s.relays.Unlock()
	for _, relay := range s.relays.m {
		if !relay.Online || relay.Noredir || relay.Disabled {
			continue
		}

		score := float64((relay.Listeners / relay.Max) - (relay.Weight / 1000))
		if score < s.min {
			s.min = score
			s.current.Store(relay.Stream)
			return
		}
	}
	s.current.Store(s.c.Fallback)
	return
}

func (s *state) check() {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	var wg sync.WaitGroup
	s.relays.Lock()
	defer s.relays.Unlock()
	for _, relay := range s.relays.m {
		if relay.Disabled {
			continue
		}
		wg.Add(1)
		go health(client, relay, &wg)
	}
	wg.Wait()
	return
}

func (s *state) getStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.relays.Lock()
		defer s.relays.Unlock()
		b, err := json.Marshal(s.relays.m)
		if err != nil {
			http.Error(w, "error marshalling relay map", http.StatusTeapot)
			return
		}
		fmt.Fprintln(w, string(b))
		return
	}
}

func (s *state) getIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, time.Now())
		return
	}
}

func (s *state) getMain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s.current.Load().(string), http.StatusFound)
		return
	}
}

func (s *state) start() {
	s.current.Store(s.c.Fallback)
	s.serv = &http.Server{
		Addr:         s.c.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	http.HandleFunc("/", s.getIndex())
	http.HandleFunc("/status", s.getStatus())
	http.HandleFunc("/main", s.getMain())

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				s.check()
				s.choose()
			case <-c:
				fmt.Println("shutting down")
				s.serv.Close()
			}
		}
	}()
	err := s.serv.ListenAndServe()
	if err != http.ErrServerClosed {
		fmt.Println("could not close gracefully", err)
	}
}
