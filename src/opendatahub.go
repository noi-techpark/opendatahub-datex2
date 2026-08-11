// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type openDataHubItem struct {
	Id          string                       `json:"Id"`
	TagIds      []string                     `json:"TagIds"`
	StartTime   time.Time                    `json:"StartTime"`
	EndTime     *time.Time                   `json:"EndTime"`
	FirstImport time.Time                    `json:"FirstImport"`
	LastChange  time.Time                    `json:"LastChange"`
	Meta        openDataHubMeta              `json:"_Meta"`
	Geo         map[string]openDataHubGeo    `json:"Geo"`
	Mapping     map[string]map[string]string `json:"Mapping"`
	Detail      map[string]openDataHubDetail `json:"Detail"`
}

type openDataHubMeta struct {
	Source     string                `json:"Source"`
	UpdateInfo openDataHubUpdateInfo `json:"UpdateInfo"`
}

type openDataHubUpdateInfo struct {
	UpdateHistory []struct{} `json:"UpdateHistory"`
}

type openDataHubGeo struct {
	Latitude  *float64 `json:"Latitude"`
	Longitude *float64 `json:"Longitude"`
}

type openDataHubDetail struct {
	Title    string `json:"Title"`
	BaseText string `json:"BaseText"`
}

type openDataHubResponse struct {
	TotalResults int               `json:"TotalResults"`
	Items        []openDataHubItem `json:"Items"`
}

func fetchEvents(url string) ([]openDataHubItem, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s from %s", resp.Status, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseEvents(body)
}

func parseEvents(body []byte) ([]openDataHubItem, error) {
	var parsed openDataHubResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return parsed.Items, nil
}

var trafficEventTags = []string{
	"traffic-event:accident",
	"traffic-event:animal-on-road",
	"traffic-event:caution",
	"traffic-event:closure",
	"traffic-event:congestion",
	"traffic-event:event",
	"traffic-event:hindrance",
	"traffic-event:prohibition",
	"traffic-event:road-condition",
	"traffic-event:road-work",
	"traffic-event:speed-camera",
	"traffic-event:weather-related",
}

var shortLivedTags = map[string]bool{
	"traffic-event:speed-camera":    true,
	"traffic-event:congestion":      true,
	"traffic-event:accident":        true,
	"traffic-event:event":           true,
	"traffic-event:animal-on-road":  true,
	"traffic-event:weather-related": true,
}

func sameLocalDay(a, b time.Time) bool {
	ay, am, ad := a.Local().Date()
	by, bm, bd := b.Local().Date()
	return ay == by && am == bm && ad == bd
}

func categoryOf(item openDataHubItem) (string, bool) {
	for _, tag := range item.TagIds {
		for _, known := range trafficEventTags {
			if tag == known {
				return tag, true
			}
		}
	}
	return "", false
}

func filterBySource(items []openDataHubItem, source string) []openDataHubItem {
	out := items[:0:0]
	for _, it := range items {
		if it.Meta.Source == source {
			out = append(out, it)
		}
	}
	return out
}

func filterCurrent(items []openDataHubItem, now time.Time) []openDataHubItem {
	out := items[:0:0]

	for _, it := range items {
		shortLived := false
		for _, tag := range it.TagIds {
			if shortLivedTags[tag] {
				shortLived = true
				break
			}
		}

		var current bool
		if shortLived {
			if it.EndTime != nil {
				current = it.EndTime.After(now)
			} else {
				current = sameLocalDay(it.StartTime, now)
			}
		} else {
			if it.EndTime == nil {
				current = true
			} else {
				current = !now.Before(it.StartTime) && !now.After(*it.EndTime)
			}
		}

		if current {
			out = append(out, it)
		}
	}
	return out
}
