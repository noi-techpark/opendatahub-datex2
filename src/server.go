// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	_ "embed"
	"net/http"
	"sync"
)

//go:embed docs/openapi.yaml
var openapiSpec []byte

//go:embed docs/redoc.html
var redocPage []byte

// Server keeps the latest rendered DATEX II XML for each recipient in
// memory and serves it at that recipient's configured path.
type Server struct {
	mu    sync.RWMutex
	pages map[string][]byte
}

func NewServer() *Server {
	return &Server{pages: make(map[string][]byte)}
}

func (s *Server) publish(path string, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pages[path] = body
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(redocPage)
		return
	case "/openapi.yaml":
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		w.Write(openapiSpec)
		return
	}

	s.mu.RLock()
	body, ok := s.pages[r.URL.Path]
	s.mu.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Write(body)
}
