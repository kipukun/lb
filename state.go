package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rivo/tview"

	"github.com/BurntSushi/toml"
)

type config struct {
	Addr, Fallback string
	Reload         bool
}

type state struct {
	serv      *http.Server
	c         config
	ui        *tview.Application
	listeners int
	min       float64
	current   atomic.Value
	relays    relays
	mtime     time.Time
}

func newState() *state {
	s := new(state)
	s.mtime = time.Now()
	s.current.Store(s.c.Fallback)
	s.serv = &http.Server{
		Addr:         s.c.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	http.HandleFunc("/", s.getIndex())
	http.HandleFunc("/status", s.getStatus())
	http.HandleFunc("/main", s.getMain())
	return s
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

func (s *state) reload() {
	f, err := os.Stat("relays.toml")
	if err != nil {
		fmt.Println("reload: error reloading from config file, disabling hot reload")
		s.c.Reload = false
	}
	if !f.ModTime().After(s.mtime) {
		return
	}
	if _, err := toml.DecodeFile("relays.toml", &s.relays.m); err != nil {
		log.Println("reload: error decoding relay config file", err)
		return
	}
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

func (s *state) start() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				if s.c.Reload {
					s.reload()
				}
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
