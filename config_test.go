package main

import (
	"strings"
	"testing"
)

func TestLoadConfig_Valid(t *testing.T) {
	input := `{
		"data_source": "./users.csv",
		"matchers": [
			{"input_field": "user", "lookup_field": "username", "method": "exact", "case_sensitive": false},
			{"input_field": "ip", "lookup_field": "ip_range", "method": "cidr"}
		]
	}`
	cfg, err := LoadConfig(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DataSource != "./users.csv" {
		t.Errorf("expected data_source=./users.csv, got %s", cfg.DataSource)
	}
	if len(cfg.Matchers) != 2 {
		t.Fatalf("expected 2 matchers, got %d", len(cfg.Matchers))
	}
	if cfg.Matchers[0].Method != "exact" {
		t.Errorf("expected method=exact, got %s", cfg.Matchers[0].Method)
	}
}

func TestLoadConfig_DefaultMethod(t *testing.T) {
	input := `{"data_source": "x.csv", "matchers": [{"input_field": "a", "lookup_field": "b"}]}`
	cfg, err := LoadConfig(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Matchers[0].Method != "exact" {
		t.Errorf("expected default method=exact, got %s", cfg.Matchers[0].Method)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	_, err := LoadConfig(strings.NewReader(`{invalid`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindMatcher_Found(t *testing.T) {
	cfg := &Config{Matchers: []Matcher{{InputField: "user"}, {InputField: "host"}}}
	m, err := cfg.FindMatcher("host")
	if err != nil {
		t.Fatal(err)
	}
	if m.InputField != "host" {
		t.Errorf("expected host, got %s", m.InputField)
	}
}

func TestFindMatcher_NotFound(t *testing.T) {
	cfg := &Config{Matchers: []Matcher{{InputField: "user"}}}
	_, err := cfg.FindMatcher("missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseMapping_WithOutput(t *testing.T) {
	m, err := ParseMapping("user as user_id OUTPUT department as dept, role")
	if err != nil {
		t.Fatal(err)
	}
	if m.ConfigRefField != "user" {
		t.Errorf("expected ConfigRefField=user, got %s", m.ConfigRefField)
	}
	if m.InputField != "user_id" {
		t.Errorf("expected InputField=user_id, got %s", m.InputField)
	}
	if m.OutputMap["department"] != "dept" {
		t.Errorf("expected department→dept, got %s", m.OutputMap["department"])
	}
	if m.OutputMap["role"] != "role" {
		t.Errorf("expected role→role, got %s", m.OutputMap["role"])
	}
}

func TestParseMapping_WithoutOutput(t *testing.T) {
	m, err := ParseMapping("hostname as hostname")
	if err != nil {
		t.Fatal(err)
	}
	if m.ConfigRefField != "hostname" {
		t.Errorf("expected hostname, got %s", m.ConfigRefField)
	}
	if len(m.OutputMap) != 0 {
		t.Errorf("expected empty OutputMap, got %v", m.OutputMap)
	}
}

func TestParseMapping_Invalid(t *testing.T) {
	_, err := ParseMapping("invalid format")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseMapping_EmptyOutputClause(t *testing.T) {
	m, err := ParseMapping("user as uid OUTPUT ")
	if err != nil {
		t.Fatal(err)
	}
	if len(m.OutputMap) != 0 {
		t.Errorf("expected empty OutputMap for empty OUTPUT clause, got %v", m.OutputMap)
	}
}
