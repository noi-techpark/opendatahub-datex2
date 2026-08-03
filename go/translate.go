// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const dateTimeLayout = "2006-01-02T15:04:05.000Z07:00"

// event mirrors the useful subset of Datex2Event / Worker.MapItemsToDatex2.
// Several fields the legacy Worker computes (SourceCountry, SourceType,
// SourceReliable, Confidentiality, InformationStatus, Urgency, Probability,
// AreaOfInterest, GeneralPublicCommentType) are never actually read by
// TraduttoreSituation.cs - it hardcodes its own values instead - so they're
// dropped here rather than carried through unused.
type event struct {
	Id           string
	Reference    string
	Version      int
	Category     string
	CreationTime time.Time
	VersionTime  time.Time
	StartTime    time.Time
	EndTime      *time.Time
	Latitude     float64
	Longitude    float64
	CommentIt    string
	CommentDe    string
	SourceName   string
}

// mapEvents mirrors Worker.MapItemsToDatex2: builds the internal event list
// from filtered ODH items, skipping any item whose tags don't match one of
// the 12 known traffic-event categories.
func mapEvents(items []odhItem, cfg *Config) []event {
	var out []event
	for _, item := range items {
		category, ok := categoryOf(item)
		if !ok {
			continue
		}

		e := event{
			Id:           strings.TrimPrefix(item.Id, cfg.IDPrefix),
			Version:      1,
			Category:     category,
			CreationTime: item.CreationTime,
			VersionTime:  item.VersionTime,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
			SourceName:   item.Meta.Source,
		}
		if n := len(item.Meta.UpdateInfo.UpdateHistory); n > 0 {
			e.Version = n
		}

		if provider, ok := item.Mapping["ProviderProvinceBz"]; ok {
			e.Reference = provider["Id"]
		}
		if geo, ok := item.Geo["position"]; ok {
			if geo.Latitude != nil {
				e.Latitude = *geo.Latitude
			}
			if geo.Longitude != nil {
				e.Longitude = *geo.Longitude
			}
		}
		if d, ok := item.Detail["it"]; ok {
			e.CommentIt = joinTitleAndText(d)
		}
		if d, ok := item.Detail["de"]; ok {
			e.CommentDe = joinTitleAndText(d)
		}

		out = append(out, e)
	}
	return out
}

func joinTitleAndText(d odhDetail) string {
	if d.BaseText == "" {
		return d.Title
	}
	return d.Title + " " + d.BaseText
}

// buildPublication mirrors TraduttoreSituation.GetD2LogicalModel. Today
// Situazione == Id always, so the legacy multi-record-per-situation grouping
// never actually groups anything - this builds one Situation with one
// SituationRecord per event, which is an equivalent simplification.
func buildPublication(events []event, cfg *Config, now time.Time) *D2LogicalModel {
	creator := InternationalIdentifier{Country: "it", NationalIdentifier: cfg.InternalSupplier}

	pub := SituationPublication{
		XsiType:            "SituationPublication",
		Lang:               "it",
		PublicationTime:    now.Format(dateTimeLayout),
		PublicationCreator: creator,
	}

	// Mirrors the legacy sort by Situazione (== Id) before translation;
	// since Id is a string, this is a lexicographic, not numeric, sort.
	sorted := make([]event, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Id < sorted[j].Id })

	for _, e := range sorted {
		subtype := cfg.subtypeFor(e.Category)
		record := buildRecord(e, subtype, cfg.InternalSupplier, now)
		if record == nil {
			continue
		}

		severity := "unknown"
		if subtype != nil && subtype.Severity != "" {
			severity = subtype.Severity
		}

		pub.Situations = append(pub.Situations, Situation{
			Id:              e.Id,
			Version:         "1",
			OverallSeverity: severity,
			HeaderInformation: HeaderInformation{
				Confidentiality:   "noRestriction",
				InformationStatus: "real",
				Urgency:           "normalUrgency",
			},
			SituationRecords: []SituationRecord{*record},
		})
	}

	return &D2LogicalModel{
		Xmlns:            "http://datex2.eu/schema/2/2_0",
		XmlnsXsi:         "http://www.w3.org/2001/XMLSchema-instance",
		ModelBaseVersion: "2",
		Exchange:         Exchange{SupplierIdentification: creator},
		Payload:          pub,
	}
}

// buildRecord mirrors TraduttoreSituation.SetSituationRecord and its
// GetXxx/SetSituationRecord_Xxx helpers. Returns nil if the event's subtype
// isn't configured/enabled, or isn't one of the known classnames (mirrors
// "Sottotipo non tradotto").
func buildRecord(e event, subtype *Subtype, internalSupplier string, now time.Time) *SituationRecord {
	if subtype == nil {
		return nil
	}

	r := &SituationRecord{
		Id:                               e.Id,
		Version:                          fmt.Sprint(e.Version),
		SituationRecordCreationReference: e.Reference,
		SituationRecordCreationTime:      e.CreationTime.Format(dateTimeLayout),
		SituationRecordVersionTime:       e.VersionTime.Local().Format(dateTimeLayout),
		ProbabilityOfOccurrence:          "certain",
		Source: Source{
			SourceType:           "roadAuthorities",
			Reliable:             true,
			SourceCountry:        "it",
			SourceIdentification: internalSupplier,
			SourceName:           MultilingualString{Values: []MultilingualStringValue{{Lang: "it", Value: e.SourceName}}},
		},
		GeneralPublicComment: []Comment{{
			Comment: MultilingualString{Values: []MultilingualStringValue{
				{Lang: "it", Value: e.CommentIt},
				{Lang: "de", Value: e.CommentDe},
			}},
			CommentDateTime: e.VersionTime.Local().Format(dateTimeLayout),
			CommentType:     "description",
		}},
		GroupOfLocations: Point{XsiType: "Point", PointByCoordinates: PointByCoordinates{PointCoordinates: PointCoordinates{
			Latitude:  float32(e.Latitude),
			Longitude: float32(e.Longitude),
		}}},
	}

	setValidity(r, e, now)

	switch subtype.Classname {
	case "SpeedManagement":
		// Fixed: the legacy GetSpeedManagement sets speedManagementType but
		// never sets the companion speedManagementTypeSpecified flag the
		// schema requires for it to be serialized - confirmed against the
		// real golden XML, where speedManagementType is silently absent
		// despite the C# code assigning it. Same class of bug as the
		// dropped headerInformation.urgency (see HeaderInformation).
		r.XsiType = "SpeedManagement"
		r.ComplianceOption = "mandatory"
		if subtype.TypeValue != "" {
			r.SpeedManagementType = subtype.TypeValue
		}
	case "MaintenanceWorks":
		r.XsiType = "MaintenanceWorks"
		if subtype.TypeValue != "" {
			r.RoadMaintenanceType = []string{subtype.TypeValue}
		}
	case "AbnormalTraffic":
		r.XsiType = "AbnormalTraffic"
		if subtype.TypeValue != "" {
			r.AbnormalTrafficType = subtype.TypeValue
		}
	case "RoadOrCarriagewayOrLaneManagement":
		r.XsiType = "RoadOrCarriagewayOrLaneManagement"
		r.ComplianceOption = "mandatory"
		if subtype.TypeValue != "" {
			r.RoadOrCarriagewayOrLaneManagementType = subtype.TypeValue
		}
	case "GeneralObstruction":
		r.XsiType = "GeneralObstruction"
		if subtype.TypeValue != "" {
			r.ObstructionType = []string{subtype.TypeValue}
		}
	case "AnimalPresenceObstruction":
		// Fixed: the legacy translator never recognized this classname
		// (seed data uses it, but the C# if/else chain only checked
		// "GeneralObstruction"), so animal-on-road events were silently
		// dropped from every publication. animalPresenceType is a single
		// required field on this DATEX type, not a list.
		r.XsiType = "AnimalPresenceObstruction"
		r.AnimalPresenceType = subtype.TypeValue
	case "WinterDrivingManagement":
		r.XsiType = "WinterDrivingManagement"
		r.ComplianceOption = "advisory"
		if subtype.TypeValue != "" {
			r.WinterEquipmentManagementType = subtype.TypeValue
		}
		if subtype.ExtraAttribute == "complianceOption" && subtype.ExtraValue != "" {
			r.ComplianceOption = subtype.ExtraValue
		}
	case "PoorEnvironmentConditions", "PoorEnvironmentCondition":
		// Fixed: seed data uses the singular "PoorEnvironmentCondition",
		// which the legacy if/else chain (checking the plural form) never
		// matched, silently dropping weather-related events too.
		r.XsiType = "PoorEnvironmentConditions"
		if subtype.TypeValue != "" {
			r.PoorEnvironmentType = []string{subtype.TypeValue}
		}
	case "Accident":
		r.XsiType = "Accident"
		if subtype.TypeValue != "" {
			r.AccidentType = []string{subtype.TypeValue}
		}
	case "PublicEvent":
		r.XsiType = "PublicEvent"
		if subtype.TypeValue != "" {
			r.PublicEventType = subtype.TypeValue
		}
	default:
		return nil
	}

	return r
}

// setValidity mirrors SetSituationRecord_ValidityTimeSpec, fixed: the
// legacy code compared a nullable EndTime with > / <= against DateTime.Now,
// which is always false when EndTime is nil - so ongoing/permanent events
// never got validityStatus=active. Here a nil EndTime is treated as "no
// upper bound" the same way the rest of the pipeline already treats it.
func setValidity(r *SituationRecord, e event, now time.Time) {
	v := Validity{
		ValidityTimeSpecification: OverallPeriod{
			OverallStartTime: e.StartTime.Format(dateTimeLayout),
		},
	}

	switch {
	case e.EndTime == nil:
		if !e.StartTime.After(now) {
			v.ValidityStatus = "active"
		} else {
			v.ValidityStatus = "definedByValidityTimeSpec"
			r.ProbabilityOfOccurrence = "probable"
		}
	case !e.StartTime.After(now) && e.EndTime.After(now):
		v.ValidityStatus = "active"
	case e.StartTime.After(now):
		v.ValidityStatus = "definedByValidityTimeSpec"
		r.ProbabilityOfOccurrence = "probable"
	default: // EndTime <= now
		v.ValidityStatus = "definedByValidityTimeSpec"
		v.Overrunning = true
	}

	if e.EndTime != nil {
		v.ValidityTimeSpecification.OverallEndTime = e.EndTime.Format(dateTimeLayout)
	}

	r.Validity = v
}
