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

func TestEquivalenceAgainstLegacyGolden(t *testing.T) {
	const (
		configPath   = "../equivalence/config.yaml"
		subtypesPath = "../equivalence/subtypes.yaml"
		fixturePath  = "../equivalence/fixtures/v1/Announcement"
		goldenPath   = "../equivalence/golden/IT492/SituationPublication.xml"
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
	provider := cfg.providerByName("province-bz")
	if provider == nil {
		t.Fatalf("config has no province-bz provider")
	}
	subtypes, err := LoadSubtypes(subtypesPath)
	if err != nil {
		t.Fatalf("load subtypes: %v", err)
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
	items = filterBySource(items, provider.Source)
	items = filterCurrent(items, now)
	events := mapEvents(items, provider)
	got := buildPublication(events, provider, subtypes, now)

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

func TestNoSilentlyEmptyFields(t *testing.T) {
	const fixturePath = "../equivalence/fixtures/v1/Announcement"
	if _, err := os.Stat(fixturePath); err != nil {
		t.Skipf("%s not present - run equivalence/run_equivalence_test.sh first", fixturePath)
	}

	cfg, err := LoadConfig("../equivalence/config.yaml")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	provider := cfg.providerByName("province-bz")
	if provider == nil {
		t.Fatalf("config has no province-bz provider")
	}
	subtypes, err := LoadSubtypes("../equivalence/subtypes.yaml")
	if err != nil {
		t.Fatalf("load subtypes: %v", err)
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
		raw[strings.TrimPrefix(it.Id, provider.IDPrefix)] = it
	}

	now := time.Now()
	filtered := filterCurrent(filterBySource(items, provider.Source), now)
	events := mapEvents(filtered, provider)
	got := buildPublication(events, provider, subtypes, now)

	checked := 0
	for _, s := range got.Payload.Situations {
		src, ok := raw[s.Id]
		if !ok || len(s.SituationRecords) != 1 {
			continue
		}
		r := s.SituationRecords[0]
		checked++

		if !src.FirstImport.IsZero() && r.SituationRecordCreationTime == "" {
			t.Errorf("situation %s: source FirstImport=%v but situationRecordCreationTime is empty", s.Id, src.FirstImport)
		}
		if !src.LastChange.IsZero() && r.SituationRecordVersionTime == "" {
			t.Errorf("situation %s: source LastChange=%v but situationRecordVersionTime is empty", s.Id, src.LastChange)
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

func compareSituation(t *testing.T, id string, got, want Situation) {
	t.Helper()

	if got.OverallSeverity != want.OverallSeverity {
		t.Errorf("situation %s: overallSeverity: got %q, want %q", id, got.OverallSeverity, want.OverallSeverity)
	}
	if got.HeaderInformation.Confidentiality != want.HeaderInformation.Confidentiality ||
		got.HeaderInformation.InformationStatus != want.HeaderInformation.InformationStatus {
		t.Errorf("situation %s: headerInformation: got %+v, want %+v", id, got.HeaderInformation, want.HeaderInformation)
	}
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
