// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/xml"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// This test proves the Go pipeline produces DATEX content equivalent to the
// real, dockerized C# app's output, for the same real Open Data Hub input.
// It does not run as part of a plain `go test ./...`: both
// ../equivalence/fixtures/v1/Announcement and
// ../equivalence/golden/IT492/SituationPublication.xml are real captures
// produced by equivalence/run_equivalence_test.sh, not checked into git (a
// real sample of short-lived event categories like "accident" is stale
// within hours - freezing one into a golden file would make the test rot on
// its own). If those files aren't present, this test skips.
//
// Because both sides read real, unpredictable data, situations are matched
// by id rather than by a fixed expected set, and exactly three kinds of
// difference are tolerated, by agreement - all bugs in the legacy app found
// by running it for real against real data:
//  1. animal-on-road and weather-related: the legacy TraduttoreSituation.cs
//     never recognized "AnimalPresenceObstruction"/"PoorEnvironmentCondition"
//     classnames, silently dropping those events entirely.
//  2. speedManagementType, and headerInformation.urgency on every
//     situation: legacy code sets the value but never sets the companion
//     Specified flag the schema needs to actually serialize it, so both are
//     silently empty/absent in the legacy output but populated in Go.
//  3. Short-lived, open-ended events (congestion/accident/event/
//     speed-camera/animal-on-road/weather-related with no EndTime): legacy
//     Worker.FiltraEventiAttuali requires ev.StartTime.Date == DateTime.Now
//     .Date without normalizing timezones first - StartTime keeps whatever
//     Kind the JSON "Z" suffix deserialized it as (UTC) while DateTime.Now
//     is server-local, so an event that started "today" in local time but
//     is still "yesterday" in UTC (the ~2h window after local midnight, for
//     Europe/Rome) is wrongly excluded. Go's sameLocalDay normalizes both
//     sides to local time first and doesn't have this bug.
//
// Any other difference is a real bug and fails the test.
func TestEquivalenceAgainstLegacyGolden(t *testing.T) {
	const (
		configPath  = "../equivalence/config.yaml"
		fixturePath = "../equivalence/fixtures/v1/Announcement"
		goldenPath  = "../equivalence/golden/IT492/SituationPublication.xml"
	)
	for _, p := range []string{fixturePath, goldenPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("%s not present - run equivalence/run_equivalence_test.sh first", p)
		}
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	items, err := parseEvents(fixture)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	goldenBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var golden D2LogicalModel
	if err := xml.Unmarshal(goldenBytes, &golden); err != nil {
		t.Fatalf("parse golden: %v", err)
	}

	now := time.Now()
	items = filterBySource(items, cfg.Source)
	items = filterCurrent(items, now)
	events := mapEvents(items, cfg)
	got := buildPublication(events, cfg, now)

	// Sanity check the Go side actually renders to well-formed XML.
	rendered, err := renderXML(got)
	if err != nil {
		t.Fatalf("render xml: %v", err)
	}
	var roundTripped D2LogicalModel
	if err := xml.Unmarshal(rendered, &roundTripped); err != nil {
		t.Fatalf("rendered XML doesn't parse: %v", err)
	}

	// Top-level fields.
	if got.Exchange.SupplierIdentification != golden.Exchange.SupplierIdentification {
		t.Errorf("exchange.supplierIdentification: got %+v, golden %+v", got.Exchange.SupplierIdentification, golden.Exchange.SupplierIdentification)
	}
	if got.Payload.PublicationCreator != golden.Payload.PublicationCreator {
		t.Errorf("publicationCreator: got %+v, golden %+v", got.Payload.PublicationCreator, golden.Payload.PublicationCreator)
	}
	if got.Payload.Lang != golden.Payload.Lang {
		t.Errorf("lang: got %q, golden %q", got.Payload.Lang, golden.Payload.Lang)
	}

	gotByID := indexByID(got.Payload.Situations)
	goldenByID := indexByID(golden.Payload.Situations)

	// isDocumentedDrop reports whether s being present only in the Go output
	// is one of the two documented legacy bugs: never-recognized classnames
	// (always dropped, regardless of dates), or the open-ended-short-lived
	// timezone bug (only for the affected classnames, and only when the
	// event actually has no end time - see the package doc comment).
	openEndedShortLived := map[string]bool{
		"AbnormalTraffic": true, // congestion
		"Accident":        true,
		"PublicEvent":     true, // event
		"SpeedManagement": true, // speed-camera
	}
	isDocumentedDrop := func(s Situation) bool {
		if len(s.SituationRecords) != 1 {
			return false
		}
		r := s.SituationRecords[0]
		switch {
		case r.XsiType == "AnimalPresenceObstruction", r.XsiType == "PoorEnvironmentConditions":
			return true
		case openEndedShortLived[r.XsiType]:
			return r.Validity.ValidityTimeSpecification.OverallEndTime == ""
		default:
			return false
		}
	}

	for id, want := range goldenByID {
		have, ok := gotByID[id]
		if !ok {
			t.Errorf("situation %s present in golden but missing from Go output", id)
			continue
		}
		compareSituation(t, id, have, want)
	}

	goOnly := 0
	for id, s := range gotByID {
		if _, inGolden := goldenByID[id]; inGolden {
			continue
		}
		goOnly++
		if !isDocumentedDrop(s) {
			t.Errorf("situation %s (xsi:type=%q) present in Go output but not in golden, and isn't a documented legacy drop",
				id, situationXsiType(s))
		}
	}

	t.Logf("compared %d golden situations against %d Go situations (%d Go-only, expected to be documented legacy drops)",
		len(goldenByID), len(gotByID), goOnly)
}

// TestNoSilentlyEmptyFields is a defense-in-depth check, independent of the
// golden comparison above: for every situation the Go pipeline keeps, it
// verifies fields the source data actually populated didn't end up
// null/empty in the output. This is exactly the failure mode of the
// CreationTime/VersionTime bug (src/opendatahub.go's history) that
// motivated this equivalence suite in the first place - it's checked
// directly against the real fetched source here rather than only relying on
// the legacy app also getting it right.
func TestNoSilentlyEmptyFields(t *testing.T) {
	const fixturePath = "../equivalence/fixtures/v1/Announcement"
	if _, err := os.Stat(fixturePath); err != nil {
		t.Skipf("%s not present - run equivalence/run_equivalence_test.sh first", fixturePath)
	}

	cfg, err := LoadConfig("../equivalence/config.yaml")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	items, err := parseEvents(fixture)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	raw := make(map[string]openDataHubItem, len(items))
	for _, it := range items {
		raw[strings.TrimPrefix(it.Id, cfg.IDPrefix)] = it
	}

	now := time.Now()
	filtered := filterCurrent(filterBySource(items, cfg.Source), now)
	events := mapEvents(filtered, cfg)
	got := buildPublication(events, cfg, now)

	checked := 0
	for _, s := range got.Payload.Situations {
		src, ok := raw[s.Id]
		if !ok || len(s.SituationRecords) != 1 {
			continue
		}
		r := s.SituationRecords[0]
		checked++

		if !src.CreationTime.IsZero() && r.SituationRecordCreationTime == "" {
			t.Errorf("situation %s: source FirstImport=%v but situationRecordCreationTime is empty", s.Id, src.CreationTime)
		}
		if !src.VersionTime.IsZero() && r.SituationRecordVersionTime == "" {
			t.Errorf("situation %s: source LastChange=%v but situationRecordVersionTime is empty", s.Id, src.VersionTime)
		}
		if geo, ok := src.Geo["position"]; ok && (geo.Latitude != nil || geo.Longitude != nil) {
			c := r.GroupOfLocations.PointByCoordinates.PointCoordinates
			if c.Latitude == 0 && c.Longitude == 0 {
				t.Errorf("situation %s: source has Geo[position]=%+v but groupOfLocations is 0,0", s.Id, geo)
			}
		}
		if d, ok := src.Detail["it"]; ok && (d.Title != "" || d.BaseText != "") && !hasCommentValue(r, "it") {
			t.Errorf("situation %s: source Detail[it]=%+v but no non-empty it comment in output", s.Id, d)
		}
		if d, ok := src.Detail["de"]; ok && (d.Title != "" || d.BaseText != "") && !hasCommentValue(r, "de") {
			t.Errorf("situation %s: source Detail[de]=%+v but no non-empty de comment in output", s.Id, d)
		}
		if src.Meta.Source != "" && r.Source.SourceName.Values == nil {
			t.Errorf("situation %s: source _Meta.Source=%q but sourceName is empty", s.Id, src.Meta.Source)
		}
	}
	t.Logf("checked %d situations against their source item for silently-empty fields", checked)
}

func hasCommentValue(r SituationRecord, lang string) bool {
	for _, c := range r.GeneralPublicComment {
		for _, v := range c.Comment.Values {
			if v.Lang == lang && v.Value != "" {
				return true
			}
		}
	}
	return false
}

func situationXsiType(s Situation) string {
	if len(s.SituationRecords) != 1 {
		return ""
	}
	return s.SituationRecords[0].XsiType
}

func indexByID(situations []Situation) map[string]Situation {
	out := make(map[string]Situation, len(situations))
	for _, s := range situations {
		out[s.Id] = s
	}
	return out
}

// compareSituation checks content equivalence field by field. Date/time
// fields are compared as instants (parsed), not as strings: the real C#
// output re-expresses timestamps in the container's local offset, which is
// the same instant in a different representation, not a mismatch.
func compareSituation(t *testing.T, id string, got, want Situation) {
	t.Helper()

	if got.OverallSeverity != want.OverallSeverity {
		t.Errorf("situation %s: overallSeverity: got %q, want %q", id, got.OverallSeverity, want.OverallSeverity)
	}
	if got.HeaderInformation.Confidentiality != want.HeaderInformation.Confidentiality ||
		got.HeaderInformation.InformationStatus != want.HeaderInformation.InformationStatus {
		t.Errorf("situation %s: headerInformation: got %+v, want %+v", id, got.HeaderInformation, want.HeaderInformation)
	}
	// Urgency is a documented divergence (see the fix note on
	// HeaderInformation in datex.go's history): legacy never serializes it.
	if got.HeaderInformation.Urgency != "normalUrgency" {
		t.Errorf("situation %s: got urgency=%q, want normalUrgency", id, got.HeaderInformation.Urgency)
	}
	if want.HeaderInformation.Urgency != "" {
		t.Errorf("situation %s: golden unexpectedly has urgency=%q - is the legacy app fixed now?", id, want.HeaderInformation.Urgency)
	}

	if len(got.SituationRecords) != 1 || len(want.SituationRecords) != 1 {
		t.Errorf("situation %s: expected exactly one situationRecord on each side, got %d vs %d", id, len(got.SituationRecords), len(want.SituationRecords))
		return
	}
	g, w := got.SituationRecords[0], want.SituationRecords[0]

	if g.XsiType != w.XsiType {
		t.Errorf("situation %s: xsi:type: got %q, want %q", id, g.XsiType, w.XsiType)
	}
	if g.Version != w.Version {
		t.Errorf("situation %s: version: got %q, want %q", id, g.Version, w.Version)
	}
	if g.SituationRecordCreationReference != w.SituationRecordCreationReference {
		t.Errorf("situation %s: creationReference: got %q, want %q", id, g.SituationRecordCreationReference, w.SituationRecordCreationReference)
	}
	if !sameInstant(t, id, "situationRecordCreationTime", g.SituationRecordCreationTime, w.SituationRecordCreationTime) {
		return
	}
	if !sameInstant(t, id, "situationRecordVersionTime", g.SituationRecordVersionTime, w.SituationRecordVersionTime) {
		return
	}
	if g.ProbabilityOfOccurrence != w.ProbabilityOfOccurrence {
		t.Errorf("situation %s: probabilityOfOccurrence: got %q, want %q", id, g.ProbabilityOfOccurrence, w.ProbabilityOfOccurrence)
	}
	if !reflect.DeepEqual(g.Source, w.Source) {
		t.Errorf("situation %s: source: got %+v, want %+v", id, g.Source, w.Source)
	}
	if g.Validity.ValidityStatus != w.Validity.ValidityStatus || g.Validity.Overrunning != w.Validity.Overrunning {
		t.Errorf("situation %s: validity status/overrunning: got %+v, want %+v", id, g.Validity, w.Validity)
	}
	if !sameInstant(t, id, "overallStartTime", g.Validity.ValidityTimeSpecification.OverallStartTime, w.Validity.ValidityTimeSpecification.OverallStartTime) {
		return
	}
	if (g.Validity.ValidityTimeSpecification.OverallEndTime == "") != (w.Validity.ValidityTimeSpecification.OverallEndTime == "") {
		t.Errorf("situation %s: overallEndTime presence differs: got %q, want %q", id, g.Validity.ValidityTimeSpecification.OverallEndTime, w.Validity.ValidityTimeSpecification.OverallEndTime)
	} else if g.Validity.ValidityTimeSpecification.OverallEndTime != "" {
		sameInstant(t, id, "overallEndTime", g.Validity.ValidityTimeSpecification.OverallEndTime, w.Validity.ValidityTimeSpecification.OverallEndTime)
	}

	if len(g.GeneralPublicComment) != 1 || len(w.GeneralPublicComment) != 1 {
		t.Errorf("situation %s: expected exactly one generalPublicComment on each side", id)
	} else if g.GeneralPublicComment[0].Comment.Values != nil && w.GeneralPublicComment[0].Comment.Values != nil {
		if len(g.GeneralPublicComment[0].Comment.Values) != len(w.GeneralPublicComment[0].Comment.Values) {
			t.Errorf("situation %s: comment values: got %+v, want %+v", id, g.GeneralPublicComment[0].Comment.Values, w.GeneralPublicComment[0].Comment.Values)
		} else {
			for i := range g.GeneralPublicComment[0].Comment.Values {
				if g.GeneralPublicComment[0].Comment.Values[i] != w.GeneralPublicComment[0].Comment.Values[i] {
					t.Errorf("situation %s: comment value %d: got %+v, want %+v", id, i, g.GeneralPublicComment[0].Comment.Values[i], w.GeneralPublicComment[0].Comment.Values[i])
				}
			}
		}
	}

	if g.GroupOfLocations != w.GroupOfLocations {
		t.Errorf("situation %s: groupOfLocations: got %+v, want %+v", id, g.GroupOfLocations, w.GroupOfLocations)
	}

	if g.ComplianceOption != w.ComplianceOption {
		t.Errorf("situation %s: complianceOption: got %q, want %q", id, g.ComplianceOption, w.ComplianceOption)
	}
	// speedManagementType is a documented divergence (see translate.go's
	// history): the legacy code never sets speedManagementTypeSpecified, so
	// it's silently dropped there but populated in Go.
	if g.XsiType == "SpeedManagement" {
		if w.SpeedManagementType != "" {
			t.Errorf("situation %s: golden unexpectedly has speedManagementType=%q - is the legacy app fixed now?", id, w.SpeedManagementType)
		}
	} else if g.SpeedManagementType != w.SpeedManagementType {
		t.Errorf("situation %s: speedManagementType: got %q, want %q", id, g.SpeedManagementType, w.SpeedManagementType)
	}
	if !equalSlice(g.RoadMaintenanceType, w.RoadMaintenanceType) {
		t.Errorf("situation %s: roadMaintenanceType: got %v, want %v", id, g.RoadMaintenanceType, w.RoadMaintenanceType)
	}
	if g.AbnormalTrafficType != w.AbnormalTrafficType {
		t.Errorf("situation %s: abnormalTrafficType: got %q, want %q", id, g.AbnormalTrafficType, w.AbnormalTrafficType)
	}
	if g.RoadOrCarriagewayOrLaneManagementType != w.RoadOrCarriagewayOrLaneManagementType {
		t.Errorf("situation %s: roadOrCarriagewayOrLaneManagementType: got %q, want %q", id, g.RoadOrCarriagewayOrLaneManagementType, w.RoadOrCarriagewayOrLaneManagementType)
	}
	if !equalSlice(g.ObstructionType, w.ObstructionType) {
		t.Errorf("situation %s: obstructionType: got %v, want %v", id, g.ObstructionType, w.ObstructionType)
	}
	if g.WinterEquipmentManagementType != w.WinterEquipmentManagementType {
		t.Errorf("situation %s: winterEquipmentManagementType: got %q, want %q", id, g.WinterEquipmentManagementType, w.WinterEquipmentManagementType)
	}
	if !equalSlice(g.PoorEnvironmentType, w.PoorEnvironmentType) {
		t.Errorf("situation %s: poorEnvironmentType: got %v, want %v", id, g.PoorEnvironmentType, w.PoorEnvironmentType)
	}
	if !equalSlice(g.AccidentType, w.AccidentType) {
		t.Errorf("situation %s: accidentType: got %v, want %v", id, g.AccidentType, w.AccidentType)
	}
	if g.PublicEventType != w.PublicEventType {
		t.Errorf("situation %s: publicEventType: got %q, want %q", id, g.PublicEventType, w.PublicEventType)
	}
}

// sameInstant compares to millisecond precision, not exact equality: Go's
// dateTimeLayout ("...Z07:00" with ".000") truncates to milliseconds, while
// the C# side preserves its native tick precision (100ns) from the same
// source timestamp - so the two are the same instant to the precision Go
// actually outputs, but never bit-for-bit equal past the millisecond. That's
// a formatting choice, not a mapping difference.
func sameInstant(t *testing.T, id, field, got, want string) bool {
	t.Helper()
	gt, err := time.Parse(time.RFC3339Nano, got)
	if err != nil {
		t.Errorf("situation %s: %s: got unparseable time %q: %v", id, field, got, err)
		return false
	}
	wt, err := time.Parse(time.RFC3339Nano, want)
	if err != nil {
		t.Errorf("situation %s: %s: want unparseable time %q: %v", id, field, want, err)
		return false
	}
	if !gt.Truncate(time.Millisecond).Equal(wt.Truncate(time.Millisecond)) {
		t.Errorf("situation %s: %s: got %q, want %q (different instants)", id, field, got, want)
		return false
	}
	return true
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
