// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"testing"
	"time"
)

// These cover date-relative filtering rules (e.g. "started today"), plus a
// regression test guarding against a single malformed item (missing the
// Geo["position"] key) breaking the entire publication cycle for every
// recipient instead of just defaulting that one item's coordinates.

func ptr(t time.Time) *time.Time { return &t }

func TestFilterCurrent_ShortLivedNoEndTimeMustStartToday(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	items := []openDataHubItem{
		{Id: "started-today", TagIds: []string{"traffic-event:congestion"}, StartTime: today},
		{Id: "started-yesterday", TagIds: []string{"traffic-event:congestion"}, StartTime: yesterday},
	}

	out := filterCurrent(items, now)
	if len(out) != 1 || out[0].Id != "started-today" {
		t.Fatalf("got %v, want only 'started-today'", ids(out))
	}
}

func TestFilterCurrent_LongPeriodNoEndTimeAlwaysCurrent(t *testing.T) {
	now := time.Now()
	items := []openDataHubItem{
		{Id: "old-caution", TagIds: []string{"traffic-event:caution"}, StartTime: now.AddDate(-2, 0, 0)},
	}
	out := filterCurrent(items, now)
	if len(out) != 1 {
		t.Fatalf("expected the open-ended long-period event to always be current, got %v", ids(out))
	}
}

func TestFilterCurrent_ExpiredEventsExcluded(t *testing.T) {
	now := time.Now()
	items := []openDataHubItem{
		{Id: "expired-long-period", TagIds: []string{"traffic-event:caution"}, StartTime: now.AddDate(0, -1, 0), EndTime: ptr(now.AddDate(0, 0, -1))},
		{Id: "expired-short-lived", TagIds: []string{"traffic-event:accident"}, StartTime: now.AddDate(0, -1, 0), EndTime: ptr(now.AddDate(0, 0, -1))},
	}
	out := filterCurrent(items, now)
	if len(out) != 0 {
		t.Fatalf("expected expired events to be excluded, got %v", ids(out))
	}
}

func TestFilterCurrent_ShortLivedFutureWithEndTimeIncluded(t *testing.T) {
	now := time.Now()
	items := []openDataHubItem{
		{Id: "future-accident", TagIds: []string{"traffic-event:accident"}, StartTime: now.AddDate(1, 0, 0), EndTime: ptr(now.AddDate(1, 0, 1))},
	}
	out := filterCurrent(items, now)
	if len(out) != 1 {
		t.Fatalf("expected a future short-lived event (with EndTime still ahead) to be included, got %v", ids(out))
	}
}

func TestFilterBySource_ExcludesOtherSources(t *testing.T) {
	items := []openDataHubItem{
		{Id: "match", Meta: openDataHubMeta{Source: "PROVINCE_BZ"}},
		{Id: "other", Meta: openDataHubMeta{Source: "OTHER_PROVINCE"}},
	}
	out := filterBySource(items, "PROVINCE_BZ")
	if len(out) != 1 || out[0].Id != "match" {
		t.Fatalf("got %v, want only 'match'", ids(out))
	}
}

func TestCategoryOf_UnknownTagSkipped(t *testing.T) {
	if _, ok := categoryOf(openDataHubItem{TagIds: []string{"traffic-event:not-a-real-category"}}); ok {
		t.Fatal("expected an unrecognized tag to not match any category")
	}
}

// Regression test for the "one bad item zeroes the whole cycle" bug: a
// missing Geo["position"] must not panic, and must not drop other events
// in the same batch - it should just default that one item's coordinates.
func TestMapEvents_MissingGeoDoesNotDropOtherEvents(t *testing.T) {
	cfg := &Config{IDPrefix: "urn:test:"}
	items := []openDataHubItem{
		{Id: "urn:test:no-geo", TagIds: []string{"traffic-event:hindrance"}},
		{Id: "urn:test:has-geo", TagIds: []string{"traffic-event:congestion"}, Geo: map[string]openDataHubGeo{
			"position": {Latitude: floatPtr(46.5), Longitude: floatPtr(11.3)},
		}},
	}

	events := mapEvents(items, cfg)
	if len(events) != 2 {
		t.Fatalf("expected both events to survive a missing Geo key on one of them, got %d", len(events))
	}

	byID := map[string]event{}
	for _, e := range events {
		byID[e.Id] = e
	}
	if e := byID["no-geo"]; e.Latitude != 0 || e.Longitude != 0 {
		t.Errorf("expected missing Geo to default to 0,0, got %v,%v", e.Latitude, e.Longitude)
	}
	if e := byID["has-geo"]; e.Latitude != 46.5 || e.Longitude != 11.3 {
		t.Errorf("expected the other event's coordinates to be unaffected, got %v,%v", e.Latitude, e.Longitude)
	}
}

func floatPtr(f float64) *float64 { return &f }

func ids(items []openDataHubItem) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Id
	}
	return out
}
