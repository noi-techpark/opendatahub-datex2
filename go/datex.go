// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

package main

import "encoding/xml"

// DATEX II 2.2.3 SituationPublication, trimmed to the elements the legacy
// TraduttoreSituation.cs actually populates. SituationRecord is normally
// polymorphic in the schema (one concrete subtype per record, selected via
// xsi:type) - here it's flattened into one struct with the handful of
// subtype-specific fields as optional, since only 10 subtypes are ever
// produced and each contributes at most two simple fields. That's simpler
// than reproducing the schema's inheritance/polymorphism machinery.

type D2LogicalModel struct {
	XMLName          xml.Name             `xml:"d2LogicalModel"`
	Xmlns            string               `xml:"xmlns,attr"`
	XmlnsXsi         string               `xml:"xmlns:xsi,attr"`
	ModelBaseVersion string               `xml:"modelBaseVersion,attr"`
	Exchange         Exchange             `xml:"exchange"`
	Payload          SituationPublication `xml:"payloadPublication"`
}

type Exchange struct {
	SupplierIdentification InternationalIdentifier `xml:"supplierIdentification"`
}

type InternationalIdentifier struct {
	Country            string `xml:"country"`
	NationalIdentifier string `xml:"nationalIdentifier"`
}

type SituationPublication struct {
	XsiType            string                  `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Lang               string                  `xml:"lang,attr"`
	PublicationTime    string                  `xml:"publicationTime"`
	PublicationCreator InternationalIdentifier `xml:"publicationCreator"`
	Situations         []Situation             `xml:"situation"`
}

type Situation struct {
	Id                string            `xml:"id,attr"`
	Version           string            `xml:"version,attr"`
	OverallSeverity   string            `xml:"overallSeverity"`
	HeaderInformation HeaderInformation `xml:"headerInformation"`
	SituationRecords  []SituationRecord `xml:"situationRecord"`
}

type HeaderInformation struct {
	Confidentiality   string `xml:"confidentiality"`
	InformationStatus string `xml:"informationStatus"`
	// Fixed: the legacy translator sets headerInformation.urgency but never
	// sets the companion urgencySpecified flag the schema requires for it
	// to actually be serialized, so it's dead output today - confirmed
	// against the real captured golden XML, which has no <urgency> element
	// despite the C# code assigning it. Same class of bug as
	// SituationRecord.SpeedManagementType below.
	Urgency string `xml:"urgency"`
}

type Source struct {
	SourceCountry        string             `xml:"sourceCountry"`
	SourceIdentification string             `xml:"sourceIdentification"`
	SourceName           MultilingualString `xml:"sourceName"`
	SourceType           string             `xml:"sourceType"`
	Reliable             bool               `xml:"reliable"`
}

type Validity struct {
	ValidityStatus            string        `xml:"validityStatus"`
	Overrunning               bool          `xml:"overrunning,omitempty"`
	ValidityTimeSpecification OverallPeriod `xml:"validityTimeSpecification"`
}

type OverallPeriod struct {
	OverallStartTime string `xml:"overallStartTime"`
	OverallEndTime   string `xml:"overallEndTime,omitempty"`
}

type Comment struct {
	Comment         MultilingualString `xml:"comment"`
	CommentDateTime string             `xml:"commentDateTime"`
	CommentType     string             `xml:"commentType"`
}

type MultilingualString struct {
	Values []MultilingualStringValue `xml:"values>value"`
}

type MultilingualStringValue struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

type Point struct {
	XsiType            string             `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	PointByCoordinates PointByCoordinates `xml:"pointByCoordinates"`
}

type PointByCoordinates struct {
	PointCoordinates PointCoordinates `xml:"pointCoordinates"`
}

type PointCoordinates struct {
	Latitude  float32 `xml:"latitude"`
	Longitude float32 `xml:"longitude"`
}

type SituationRecord struct {
	XsiType                          string    `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Id                               string    `xml:"id,attr"`
	Version                          string    `xml:"version,attr"`
	SituationRecordCreationReference string    `xml:"situationRecordCreationReference,omitempty"`
	SituationRecordCreationTime      string    `xml:"situationRecordCreationTime"`
	SituationRecordVersionTime       string    `xml:"situationRecordVersionTime"`
	Source                           Source    `xml:"source"`
	ProbabilityOfOccurrence          string    `xml:"probabilityOfOccurrence"`
	Validity                         Validity  `xml:"validity"`
	GeneralPublicComment             []Comment `xml:"generalPublicComment,omitempty"`
	GroupOfLocations                 Point     `xml:"groupOfLocations"`

	// Subtype-specific fields - only the ones matching XsiType are populated.
	SpeedManagementType                   string   `xml:"speedManagementType,omitempty"`
	ComplianceOption                      string   `xml:"complianceOption,omitempty"`
	RoadMaintenanceType                   []string `xml:"roadMaintenanceType,omitempty"`
	AbnormalTrafficType                   string   `xml:"abnormalTrafficType,omitempty"`
	RoadOrCarriagewayOrLaneManagementType string   `xml:"roadOrCarriagewayOrLaneManagementType,omitempty"`
	ObstructionType                       []string `xml:"obstructionType,omitempty"`
	WinterEquipmentManagementType         string   `xml:"winterEquipmentManagementType,omitempty"`
	AnimalPresenceType                    string   `xml:"animalPresenceType,omitempty"`
	PoorEnvironmentType                   []string `xml:"poorEnvironmentType,omitempty"`
	AccidentType                          []string `xml:"accidentType,omitempty"`
	PublicEventType                       string   `xml:"publicEventType,omitempty"`
}

func renderXML(model *D2LogicalModel) ([]byte, error) {
	out, err := xml.MarshalIndent(model, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}
