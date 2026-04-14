package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Config represents the lookup configuration file structure.
type Config struct {
	DataSource string    `json:"data_source"`
	Matchers   []Matcher `json:"matchers"`
}

// Matcher defines a single matching rule.
type Matcher struct {
	InputField    string `json:"input_field"`
	LookupField   string `json:"lookup_field"`
	Method        string `json:"method"`
	CaseSensitive bool   `json:"case_sensitive"`
}

// Mapping holds the parsed -m flag value.
type Mapping struct {
	ConfigRefField string
	InputField     string
	OutputMap      map[string]string // source field → target field name
}

// LoadConfig reads and parses a JSON config from r.
func LoadConfig(r io.Reader) (*Config, error) {
	var cfg Config
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("could not parse config JSON: %w", err)
	}
	// Default method to "exact" if omitted
	for i := range cfg.Matchers {
		if cfg.Matchers[i].Method == "" {
			cfg.Matchers[i].Method = "exact"
		}
	}
	return &cfg, nil
}

// FindMatcher returns the matcher whose InputField matches name.
func (c *Config) FindMatcher(name string) (*Matcher, error) {
	for i := range c.Matchers {
		if c.Matchers[i].InputField == name {
			return &c.Matchers[i], nil
		}
	}
	return nil, fmt.Errorf("no matcher found for '%s' in config", name)
}

// mappingRe parses: <config_ref> as <input_field> [OUTPUT <fields>]
var mappingRe = regexp.MustCompile(`^(\S+)\s+as\s+(\S+)(?:\s+OUTPUT\s*(.*))?$`)

// ParseMapping parses a mapping rule string.
func ParseMapping(s string) (*Mapping, error) {
	matches := mappingRe.FindStringSubmatch(strings.TrimSpace(s))
	if matches == nil {
		return nil, fmt.Errorf("invalid mapping format: %q (expected: '<ref> as <field> [OUTPUT <mappings>]')", s)
	}

	m := &Mapping{
		ConfigRefField: matches[1],
		InputField:     matches[2],
		OutputMap:      make(map[string]string),
	}

	if matches[3] != "" {
		for _, part := range strings.Split(matches[3], ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if idx := strings.Index(part, " as "); idx != -1 {
				src := strings.TrimSpace(part[:idx])
				dst := strings.TrimSpace(part[idx+4:])
				m.OutputMap[src] = dst
			} else {
				m.OutputMap[part] = part
			}
		}
	}

	return m, nil
}
