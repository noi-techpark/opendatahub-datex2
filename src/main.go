// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML config")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("initial config load failed: %v", err)
	}
	if len(cfg.Providers) == 0 {
		log.Fatalf("config has no providers")
	}

	srv := NewServer()
	srv.setBaseURL(cfg.BaseURL)
	go func() {
		log.Fatal(http.ListenAndServe(cfg.ListenAddr, srv))
	}()

	var wg sync.WaitGroup
	for _, provider := range cfg.Providers {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			runProviderLoop(*configPath, name, srv)
		}(provider.Name)
	}
	wg.Wait()
}

func runProviderLoop(configPath, providerName string, srv *Server) {
	for {
		cfg, err := LoadConfig(configPath)
		if err != nil {
			log.Printf("[%s] reload config: %v", providerName, err)
			time.Sleep(10 * time.Second)
			continue
		}
		srv.setBaseURL(cfg.BaseURL)

		provider := cfg.providerByName(providerName)
		if provider == nil {
			log.Printf("[%s] provider no longer in config, stopping", providerName)
			return
		}

		subtypes, err := LoadSubtypes(cfg.subtypesPath())
		if err != nil {
			log.Printf("[%s] load subtypes: %v", providerName, err)
			time.Sleep(10 * time.Second)
			continue
		}

		runCycle(provider, subtypes, srv)
		time.Sleep(time.Duration(provider.PollIntervalSeconds) * time.Second)
	}
}

func runCycle(provider *ProviderConfig, subtypes SubtypeMap, srv *Server) {
	items, err := fetchEvents(provider.EventsURL)
	if err != nil {
		log.Printf("[%s] fetch events: %v", provider.Name, err)
		return
	}

	now := time.Now()
	items = filterBySource(items, provider.Source)
	items = filterCurrent(items, now)
	events := mapEvents(items, provider)

	model := buildPublication(events, provider, subtypes, now)
	body, err := renderXML(model)
	if err != nil {
		log.Printf("[%s] render xml: %v", provider.Name, err)
		return
	}

	published := 0
	for _, rec := range provider.Recipients {
		if !rec.Enabled {
			continue
		}
		srv.publish(provider.Name, rec.Type, rec.Path, body)
		published++
	}

	log.Printf("[%s] published %d events to %d recipients", provider.Name, len(events), published)
}
