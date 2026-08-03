// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config replaces appsettings.json + the three Postgres config tables
// (TAB_DESTINATARI, TAB_PARAMETRI, TAB_SOTTOTIPI) from the legacy service.
// It is re-read on every poll cycle, mirroring the legacy behavior of
// querying Postgres fresh each time (so config changes apply without a
// restart).
type Config struct {
	ListenAddr          string      `yaml:"listenAddr"`
	EventsURL           string      `yaml:"eventsUrl"`
	Source              string      `yaml:"source"`
	PollIntervalSeconds int         `yaml:"pollIntervalSeconds"`
	IDPrefix            string      `yaml:"idPrefix"`
	InternalSupplier    string      `yaml:"internalSupplier"`
	Subtypes            []Subtype   `yaml:"subtypes"`
	Recipients          []Recipient `yaml:"recipients"`
}

// Subtype replaces a TAB_SOTTOTIPI row. TypeCode was always "GENERIC" and
// Ingresso was never read by the legacy pipeline, so both are dropped.
// Uscita+Abilitato (which were always checked together) collapse into Enabled.
type Subtype struct {
	Category       string `yaml:"category"`
	Classname      string `yaml:"classname"`
	TypeValue      string `yaml:"typeValue"`
	ExtraAttribute string `yaml:"extraAttribute"`
	ExtraValue     string `yaml:"extraValue"`
	Severity       string `yaml:"severity"`
	Enabled        bool   `yaml:"enabled"`
}

// Recipient replaces a TAB_DESTINATARI row. Invio/Ricezione/InternalSupplier/
// Traduttore/Confidentiality/Url were never read by the legacy pipeline, so
// they're dropped. Path is new: it's where this recipient's XML is served.
type Recipient struct {
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
