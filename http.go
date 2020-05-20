package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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
