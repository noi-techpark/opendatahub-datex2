// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr   string           `yaml:"listenAddr"`
	BaseURL      string           `yaml:"baseUrl"`
	SubtypesFile string           `yaml:"subtypesFile"`
	Providers    []ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Name                string      `yaml:"name"`
	EventsURL           string      `yaml:"eventsUrl"`
	Source              string      `yaml:"source"`
	PollIntervalSeconds int         `yaml:"pollIntervalSeconds"`
	IDPrefix            string      `yaml:"idPrefix"`
	InternalSupplier    string      `yaml:"internalSupplier"`
	Recipients          []Recipient `yaml:"recipients"`
}

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

func (c *Config) providerByName(name string) *ProviderConfig {
	for i := range c.Providers {
		if c.Providers[i].Name == name {
			return &c.Providers[i]
		}
	}
	return nil
}

func (c *Config) subtypesPath() string {
	if c.SubtypesFile != "" {
		return c.SubtypesFile
	}
	return "subtypes.yaml"
}
