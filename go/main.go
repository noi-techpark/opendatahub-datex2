// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML config")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("initial config load failed: %v", err)
	}

	srv := NewServer()
	go func() {
		log.Fatal(http.ListenAndServe(cfg.ListenAddr, srv))
	}()

	for {
		if reloaded, err := LoadConfig(*configPath); err != nil {
			log.Printf("reload config: %v", err)
		} else {
			cfg = reloaded
		}

		runCycle(cfg, srv)
		time.Sleep(time.Duration(cfg.PollIntervalSeconds) * time.Second)
	}
}

// runCycle mirrors one iteration of Worker.ExecuteAsync's loop: fetch,
// filter, translate, then publish to every enabled recipient.
func runCycle(cfg *Config, srv *Server) {
	items, err := fetchEvents(cfg.EventsURL)
	if err != nil {
		log.Printf("fetch events: %v", err)
		return
	}

	now := time.Now()
	items = filterBySource(items, cfg.Source)
	items = filterCurrent(items, now)
	events := mapEvents(items, cfg)

	model := buildPublication(events, cfg, now)
	body, err := renderXML(model)
	if err != nil {
		log.Printf("render xml: %v", err)
		return
	}

	published := 0
	for _, rec := range cfg.Recipients {
		if !rec.Enabled {
			continue
		}
		srv.publish(rec.Path, body)
		published++
	}

	log.Printf("published %d events to %d recipients", len(events), published)
}
