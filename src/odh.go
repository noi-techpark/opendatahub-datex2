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

// odhItem is the subset of Open Data Hub Announcement item fields this
// service reads, matching the field list requested in eventsUrl.
type odhItem struct {
	Id           string                       `json:"Id"`
	TagIds       []string                     `json:"TagIds"`
	StartTime    time.Time                    `json:"StartTime"`
	EndTime      *time.Time                   `json:"EndTime"`
	CreationTime time.Time                    `json:"CreationTime"`
	VersionTime  time.Time                    `json:"VersionTime"`
	Meta         odhMeta                      `json:"_Meta"`
	Geo          map[string]odhGeo            `json:"Geo"`
	Mapping      map[string]map[string]string `json:"Mapping"`
	Detail       map[string]odhDetail         `json:"Detail"`
}

type odhMeta struct {
	Source     string        `json:"Source"`
	UpdateInfo odhUpdateInfo `json:"UpdateInfo"`
}

type odhUpdateInfo struct {
	UpdateHistory []struct{} `json:"UpdateHistory"`
}

type odhGeo struct {
	Latitude  *float64 `json:"Latitude"`
	Longitude *float64 `json:"Longitude"`
}

type odhDetail struct {
	Title    string `json:"Title"`
	BaseText string `json:"BaseText"`
}

type odhResponse struct {
	TotalResults int       `json:"TotalResults"`
	Items        []odhItem `json:"Items"`
}

func fetchEvents(url string) ([]odhItem, error) {
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

func parseEvents(body []byte) ([]odhItem, error) {
	var parsed odhResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return parsed.Items, nil
}

// The 12 traffic-event tags this service recognizes, and the subset of
// those treated as "short-lived" for the purposes of current-event
// filtering.
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

func categoryOf(item odhItem) (string, bool) {
	for _, tag := range item.TagIds {
		for _, known := range trafficEventTags {
			if tag == known {
				return tag, true
			}
		}
	}
	return "", false
}

// filterBySource re-applies the source filter client-side as a safety net
// on top of the source= query parameter already in eventsUrl, in case the
// upstream API ever returns items from other sources. The source id to
// match against is config-driven, so this works for any provider's feed.
func filterBySource(items []odhItem, source string) []odhItem {
	out := items[:0:0]
	for _, it := range items {
		if it.Meta.Source == source {
			out = append(out, it)
		}
	}
	return out
}

// filterCurrent keeps only events that are still relevant now: short-lived
// categories need to have started today (if open-ended) or not yet ended;
// everything else is valid unless it has an end time that's already
// passed.
func filterCurrent(items []odhItem, now time.Time) []odhItem {
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
