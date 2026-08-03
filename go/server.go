// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

package main

import (
	"net/http"
	"sync"
)

// Server replaces Worker.SaveToFile: instead of writing one XML file per
// recipient to disk, it keeps the latest rendered bytes in memory and
// serves them at each recipient's configured path.
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
