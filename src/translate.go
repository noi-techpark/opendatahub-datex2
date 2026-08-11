// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const dateTimeLayout = "2006-01-02T15:04:05.000Z07:00"

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

func mapEvents(items []openDataHubItem, provider *ProviderConfig) []event {
	var out []event
	for _, item := range items {
		category, ok := categoryOf(item)
		if !ok {
			continue
		}

		e := event{
			Id:           strings.TrimPrefix(item.Id, provider.IDPrefix),
			Version:      1,
			Category:     category,
			CreationTime: item.FirstImport,
			VersionTime:  item.LastChange,
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

func joinTitleAndText(d openDataHubDetail) string {
	if d.BaseText == "" {
		return d.Title
	}
	return d.Title + " " + d.BaseText
}

func buildPublication(events []event, provider *ProviderConfig, subtypes SubtypeMap, now time.Time) *D2LogicalModel {
	creator := InternationalIdentifier{Country: "it", NationalIdentifier: provider.InternalSupplier}

	pub := SituationPublication{
		XsiType:            "SituationPublication",
		Lang:               "it",
		PublicationTime:    now.Format(dateTimeLayout),
		PublicationCreator: creator,
	}

	sorted := make([]event, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Id < sorted[j].Id })

	for _, e := range sorted {
		subtype := subtypes.subtypeFor(e.Category)
		record := buildRecord(e, subtype, provider.InternalSupplier, now)
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
		// animalPresenceType is a single required field on this DATEX type,
		// not a list.
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
		// Handles both spellings: config data uses the singular
		// "PoorEnvironmentCondition", while the canonical DATEX II
		// classname is the plural "PoorEnvironmentConditions".
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
