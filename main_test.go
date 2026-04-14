package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// compareJSON compares two JSON strings by parsing and comparing maps.
func compareJSON(t *testing.T, got, want string) {
	t.Helper()
	var gotObj, wantObj interface{}
	if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
		t.Fatalf("could not parse got JSON: %v\ngot: %s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &wantObj); err != nil {
		t.Fatalf("could not parse want JSON: %v\nwant: %s", err, want)
	}
	if !reflect.DeepEqual(gotObj, wantObj) {
		t.Errorf("JSON mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

// --- Regression tests using testdata ---

func TestRegression_ExactMatch(t *testing.T) {
	input, _ := os.ReadFile("testdata/input.jsonl")
	expected, _ := os.ReadFile("testdata/exact_match.expected.jsonl")

	var buf bytes.Buffer
	opts := options{
		configFile: "testdata/lookup_config.json",
		mappingStr: "user as user OUTPUT department as dept, role",
	}
	if err := execute(opts, bytes.NewReader(input), &buf); err != nil {
		t.Fatal(err)
	}

	gotLines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	wantLines := strings.Split(strings.TrimSpace(string(expected)), "\n")

	if len(gotLines) != len(wantLines) {
		t.Fatalf("line count mismatch: got %d, want %d", len(gotLines), len(wantLines))
	}
	for i := range gotLines {
		compareJSON(t, gotLines[i], wantLines[i])
	}
}

func TestRegression_WildcardMatch(t *testing.T) {
	input, _ := os.ReadFile("testdata/input.jsonl")
	expected, _ := os.ReadFile("testdata/wildcard_match.expected.jsonl")

	var buf bytes.Buffer
	opts := options{
		configFile: "testdata/lookup_config.json",
		mappingStr: "hostname as hostname",
	}
	if err := execute(opts, bytes.NewReader(input), &buf); err != nil {
		t.Fatal(err)
	}

	gotLines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	wantLines := strings.Split(strings.TrimSpace(string(expected)), "\n")

	if len(gotLines) != len(wantLines) {
		t.Fatalf("line count mismatch: got %d, want %d", len(gotLines), len(wantLines))
	}
	for i := range gotLines {
		compareJSON(t, gotLines[i], wantLines[i])
	}
}

func TestRegression_RegexMatch(t *testing.T) {
	input, _ := os.ReadFile("testdata/input.jsonl")
	expected, _ := os.ReadFile("testdata/regex_match.expected.jsonl")

	var buf bytes.Buffer
	opts := options{
		configFile: "testdata/lookup_config.json",
		mappingStr: "process as process",
	}
	if err := execute(opts, bytes.NewReader(input), &buf); err != nil {
		t.Fatal(err)
	}

	gotLines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	wantLines := strings.Split(strings.TrimSpace(string(expected)), "\n")

	if len(gotLines) != len(wantLines) {
		t.Fatalf("line count mismatch: got %d, want %d", len(gotLines), len(wantLines))
	}
	for i := range gotLines {
		compareJSON(t, gotLines[i], wantLines[i])
	}
}

func TestRegression_CIDRMatchArray(t *testing.T) {
	input, _ := os.ReadFile("testdata/input_array.json")
	expected, _ := os.ReadFile("testdata/cidr_match_array.expected.json")

	var buf bytes.Buffer
	opts := options{
		configFile: "testdata/lookup_config.json",
		mappingStr: "client_ip as client_ip",
	}
	if err := execute(opts, bytes.NewReader(input), &buf); err != nil {
		t.Fatal(err)
	}

	// Compare as JSON arrays
	compareJSON(t, buf.String(), string(expected))
}

// --- execute unit tests ---

func TestExecute_MissingConfig(t *testing.T) {
	err := execute(options{mappingStr: "x as x"}, strings.NewReader(""), nil)
	if err == nil || !strings.Contains(err.Error(), "-c (config file) is required") {
		t.Errorf("expected config required error, got: %v", err)
	}
}

func TestExecute_InvalidMapping(t *testing.T) {
	err := execute(options{mappingStr: "invalid"}, strings.NewReader(""), nil)
	if err == nil || !strings.Contains(err.Error(), "invalid mapping format") {
		t.Errorf("expected mapping error, got: %v", err)
	}
}

func TestExecute_ConfigNotFound(t *testing.T) {
	err := execute(options{
		configFile: "/nonexistent/config.json",
		mappingStr: "x as x",
	}, strings.NewReader(""), nil)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestExecute_EmptyInput(t *testing.T) {
	var buf bytes.Buffer
	opts := options{
		configFile: "testdata/lookup_config.json",
		mappingStr: "user as user",
	}
	if err := execute(opts, strings.NewReader(""), &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

// --- path resolution (ported from legacy tests) ---

func TestResolveDataSourcePath_Cases(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		name      string
		source    string
		configDir string
		want      string
	}{
		{"relative", "./data.csv", "/etc/lookup", "/etc/lookup/data.csv"},
		{"absolute", "/data/users.csv", "/etc/lookup", "/data/users.csv"},
		{"tilde", "~/data.csv", "/etc/lookup", home + "/data.csv"},
		{"just filename", "users.csv", "/etc/lookup", "/etc/lookup/users.csv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveDataSourcePath(tt.source, tt.configDir)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// --- generate-config test ---

func TestGenerateConfig_CSVIntegration(t *testing.T) {
	var buf bytes.Buffer
	if err := generateConfig(&buf, "testdata/users.csv"); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := json.Unmarshal(buf.Bytes(), &cfg); err != nil {
		t.Fatal(err)
	}

	// users.csv has 5 columns: username,department,role,building,ip_range
	if len(cfg.Matchers) != 5 {
		names := make([]string, len(cfg.Matchers))
		for i, m := range cfg.Matchers {
			names[i] = m.InputField
		}
		sort.Strings(names)
		t.Fatalf("expected 5 matchers for users.csv, got %d: %v", len(cfg.Matchers), names)
	}
}
