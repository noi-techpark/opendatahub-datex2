// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Subtype struct {
	Category       string `yaml:"category"`
	Classname      string `yaml:"classname"`
	TypeValue      string `yaml:"typeValue"`
	ExtraAttribute string `yaml:"extraAttribute"`
	ExtraValue     string `yaml:"extraValue"`
	Severity       string `yaml:"severity"`
	Enabled        bool   `yaml:"enabled"`
}

type SubtypeMap []Subtype

type subtypesFile struct {
	Subtypes []Subtype `yaml:"subtypes"`
}

func LoadSubtypes(path string) (SubtypeMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read subtypes: %w", err)
	}
	var f subtypesFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse subtypes: %w", err)
	}
	return SubtypeMap(f.Subtypes), nil
}

func (m SubtypeMap) subtypeFor(category string) *Subtype {
	for i := range m {
		if m[i].Category == category && m[i].Enabled {
			return &m[i]
		}
	}
	return nil
}
