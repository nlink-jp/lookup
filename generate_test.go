package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateConfig_CSV(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	os.WriteFile(csvPath, []byte("username,department,role\njdoe,Sales,Manager\n"), 0644)

	var buf bytes.Buffer
	if err := generateConfig(&buf, csvPath); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := json.Unmarshal(buf.Bytes(), &cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Matchers) != 3 {
		t.Fatalf("expected 3 matchers, got %d", len(cfg.Matchers))
	}
	if cfg.DataSource != csvPath {
		t.Errorf("expected data_source=%s, got %s", csvPath, cfg.DataSource)
	}
}

func TestGenerateConfig_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test.json")
	os.WriteFile(jsonPath, []byte(`[{"name":"a","value":1},{"name":"b","extra":true}]`), 0644)

	var buf bytes.Buffer
	if err := generateConfig(&buf, jsonPath); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	json.Unmarshal(buf.Bytes(), &cfg)
	// Should find 3 unique keys: name, value, extra
	if len(cfg.Matchers) != 3 {
		t.Fatalf("expected 3 matchers (name,value,extra), got %d", len(cfg.Matchers))
	}
}

func TestGenerateConfig_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	txtPath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(txtPath, []byte("hello"), 0644)

	var buf bytes.Buffer
	if err := generateConfig(&buf, txtPath); err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestGenerateConfig_FileNotFound(t *testing.T) {
	var buf bytes.Buffer
	if err := generateConfig(&buf, "/nonexistent/file.csv"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
