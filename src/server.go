// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"sync"
)

//go:embed docs/openapi.yaml
var openapiSpec []byte

//go:embed docs/redoc.html
var redocPage []byte

// file is one published DATEX II document, for the discoverability
// endpoint.
type file struct {
	Provider string `json:"-"`
	Type     string `json:"type"`
	Path     string `json:"path"`
}

// Server keeps the latest rendered DATEX II XML for each recipient in
// memory and serves it at that recipient's configured path.
type Server struct {
	mu      sync.RWMutex
	baseURL string
	pages   map[string][]byte
	files   map[string]file
}

func NewServer() *Server {
	return &Server{pages: make(map[string][]byte), files: make(map[string]file)}
}

func (s *Server) setBaseURL(baseURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baseURL = baseURL
}

func (s *Server) publish(provider, fileType, path string, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pages[path] = body
	s.files[path] = file{Provider: provider, Type: fileType, Path: path}
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
	case "/datex/2/":
		s.serveProviderList(w)
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

// serveProviderList answers the discoverability endpoint: which providers
// are available, and which DATEX II files each one currently publishes.
func (s *Server) serveProviderList(w http.ResponseWriter) {
	s.mu.RLock()
	baseURL := s.baseURL
	files := make([]file, 0, len(s.files))
	for _, f := range s.files {
		files = append(files, f)
	}
	s.mu.RUnlock()

	providers := map[string]map[string]string{}
	for _, f := range files {
		if providers[f.Provider] == nil {
			providers[f.Provider] = map[string]string{}
		}
		providers[f.Provider][f.Type] = baseURL + f.Path
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(providers)
}
