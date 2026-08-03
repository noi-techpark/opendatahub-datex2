// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

package main

import (
	"encoding/xml"
	"os"
	"reflect"
	"testing"
	"time"
)

// This test proves the Go pipeline produces DATEX content equivalent to the
// real, dockerized C# app's output, for the same input events. The golden
// XML in ../equivalence/golden/ was captured once by regen_golden.sh; this
// test doesn't need Docker to run.
//
// Several things are deliberately NOT equivalent, by agreement - all silent
// data-loss bugs found by comparing against the real golden output:
//  1. animal-on-road and weather-related: the legacy TraduttoreSituation.cs
//     never recognized seed data's "AnimalPresenceObstruction"/
//     "PoorEnvironmentCondition" classnames, silently dropping those events
//     entirely (two fewer situations in golden than in the Go output).
//  2. speed-camera's speedManagementType, and headerInformation.urgency on
//     every situation: legacy code sets the value but never sets the
//     companion Specified flag the schema needs to actually serialize it,
//     so both are silently empty/absent in golden but populated in Go.
//
// Everything else - the other subtypes, filtering (excluded/expired/
// unknown-category/wrong-source events), the disabled "prohibition" subtype,
// per-recipient fan-out - is expected to match.
func TestEquivalenceAgainstLegacyGolden(t *testing.T) {
	cfg, err := LoadConfig("../equivalence/config.yaml")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	fixture, err := os.ReadFile("../equivalence/fixtures/v1/Announcement")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	items, err := parseEvents(fixture)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	goldenBytes, err := os.ReadFile("../equivalence/golden/IT492/SituationPublication.xml")
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

	knownFixedIDs := map[string]bool{"3": true, "6": true} // animal-on-road, weather-related

	for id, want := range goldenByID {
		have, ok := gotByID[id]
		if !ok {
			t.Errorf("situation %s present in golden but missing from Go output", id)
			continue
		}
		compareSituation(t, id, have, want)
	}

	for id := range gotByID {
		if _, inGolden := goldenByID[id]; !inGolden && !knownFixedIDs[id] {
			t.Errorf("situation %s present in Go output but not in golden, and isn't one of the documented fixes", id)
		}
	}

	for id := range knownFixedIDs {
		if _, ok := gotByID[id]; !ok {
			t.Errorf("situation %s (documented fix) missing from Go output", id)
		}
		if _, ok := goldenByID[id]; ok {
			t.Errorf("situation %s (documented fix) unexpectedly present in golden - is the golden stale?", id)
		}
	}

	// Both fixed categories should have produced sensible records.
	if s, ok := gotByID["3"]; ok {
		rec := s.SituationRecords[0]
		if rec.XsiType != "AnimalPresenceObstruction" || rec.AnimalPresenceType != "animalsOnTheRoad" {
			t.Errorf("situation 3: got xsi:type=%q animalPresenceType=%q", rec.XsiType, rec.AnimalPresenceType)
		}
	}
	if s, ok := gotByID["6"]; ok {
		rec := s.SituationRecords[0]
		if rec.XsiType != "PoorEnvironmentConditions" || len(rec.PoorEnvironmentType) != 1 || rec.PoorEnvironmentType[0] != "fog" {
			t.Errorf("situation 6: got xsi:type=%q poorEnvironmentType=%v", rec.XsiType, rec.PoorEnvironmentType)
		}
	}

	// Third documented fix: speedManagementType (situation 11) is silently
	// dropped in the legacy output (missing Specified flag) but populated
	// in the Go output.
	if s, ok := gotByID["11"]; ok {
		if s.SituationRecords[0].SpeedManagementType != "policeSpeedChecksInOperation" {
			t.Errorf("situation 11: got speedManagementType=%q, want policeSpeedChecksInOperation", s.SituationRecords[0].SpeedManagementType)
		}
	}
	if s, ok := goldenByID["11"]; ok && s.SituationRecords[0].SpeedManagementType != "" {
		t.Errorf("situation 11: golden unexpectedly has speedManagementType=%q - is the golden stale?", s.SituationRecords[0].SpeedManagementType)
	}

	// Events that must be dropped by both implementations: disabled subtype,
	// already-expired, unknown category, wrong source.
	for _, id := range []string{"8", "13-ended-long-period", "14-ended-short-lived", "15-unknown-tag", "16-wrong-source"} {
		if _, ok := gotByID[id]; ok {
			t.Errorf("situation %s should have been dropped (disabled subtype/expired/unknown/wrong source) but is present in Go output", id)
		}
		if _, ok := goldenByID[id]; ok {
			t.Errorf("situation %s should have been dropped but is present in golden - fixture/golden mismatch", id)
		}
	}

	if len(goldenByID)+len(knownFixedIDs) != len(gotByID) {
		t.Errorf("situation count: golden=%d + documented fixes=%d, want got=%d, have got=%d",
			len(goldenByID), len(knownFixedIDs), len(goldenByID)+len(knownFixedIDs), len(gotByID))
	}
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
// output re-expresses timestamps in the container's local offset (e.g. a
// fixture time given as "+01:00" comes back as "+02:00" once DST applies),
// which is the same instant in a different representation, not a mismatch.
func compareSituation(t *testing.T, id string, got, want Situation) {
	t.Helper()

	if got.OverallSeverity != want.OverallSeverity {
		t.Errorf("situation %s: overallSeverity: got %q, want %q", id, got.OverallSeverity, want.OverallSeverity)
	}
	if got.HeaderInformation.Confidentiality != want.HeaderInformation.Confidentiality ||
		got.HeaderInformation.InformationStatus != want.HeaderInformation.InformationStatus {
		t.Errorf("situation %s: headerInformation: got %+v, want %+v", id, got.HeaderInformation, want.HeaderInformation)
	}
	// Urgency is intentionally not compared against golden: fixed dead field,
	// see the fix note on HeaderInformation in datex.go. Checked below instead.
	if got.HeaderInformation.Urgency != "normalUrgency" {
		t.Errorf("situation %s: got urgency=%q, want normalUrgency", id, got.HeaderInformation.Urgency)
	}
	if want.HeaderInformation.Urgency != "" {
		t.Errorf("situation %s: golden unexpectedly has urgency=%q - is the golden stale?", id, want.HeaderInformation.Urgency)
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
	// speedManagementType is intentionally not compared against golden: the
	// legacy code never sets speedManagementTypeSpecified, so it's silently
	// dropped there (see the fix note in translate.go). Checked separately
	// below for situation 11.
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
	if !gt.Equal(wt) {
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
