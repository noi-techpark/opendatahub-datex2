// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is loaded from a YAML file and re-read on every poll cycle, so
// config changes apply without a restart.
type Config struct {
	ListenAddr          string      `yaml:"listenAddr"`
	BaseURL             string      `yaml:"baseUrl"`
	Provider            string      `yaml:"provider"`
	EventsURL           string      `yaml:"eventsUrl"`
	Source              string      `yaml:"source"`
	PollIntervalSeconds int         `yaml:"pollIntervalSeconds"`
	IDPrefix            string      `yaml:"idPrefix"`
	InternalSupplier    string      `yaml:"internalSupplier"`
	Subtypes            []Subtype   `yaml:"subtypes"`
	Recipients          []Recipient `yaml:"recipients"`
}

// Subtype maps one Open Data Hub traffic-event category to the DATEX II
// classname/typeValue it should be published as.
type Subtype struct {
	Category       string `yaml:"category"`
	Classname      string `yaml:"classname"`
	TypeValue      string `yaml:"typeValue"`
	ExtraAttribute string `yaml:"extraAttribute"`
	ExtraValue     string `yaml:"extraValue"`
	Severity       string `yaml:"severity"`
	Enabled        bool   `yaml:"enabled"`
}

// Recipient is one DATEX II document published for this provider. Type
// identifies which document it is (e.g. "situation-publication") - a
// provider may publish more than one DATEX II document type in the future,
// and not every provider necessarily offers the same set. Path is where it
// is served; by convention it should follow
// /datex/2/{provider}/{type}.xml.
type Recipient struct {
	Type        string `yaml:"type"`
	Supplier    string `yaml:"supplier"`
	Description string `yaml:"description"`
	Path        string `yaml:"path"`
	Enabled     bool   `yaml:"enabled"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) subtypeFor(category string) *Subtype {
	for i := range c.Subtypes {
		if c.Subtypes[i].Category == category && c.Subtypes[i].Enabled {
			return &c.Subtypes[i]
		}
	}
	return nil
}
