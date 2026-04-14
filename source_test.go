package main

import (
	"strings"
	"testing"
)

func TestLoadCSV_Basic(t *testing.T) {
	input := "username,department,role\njdoe,Sales,Manager\nasmith,Engineering,Developer\n"
	data, err := LoadCSV(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(data))
	}
	if data[0]["username"] != "jdoe" {
		t.Errorf("expected jdoe, got %s", data[0]["username"])
	}
	if data[1]["department"] != "Engineering" {
		t.Errorf("expected Engineering, got %s", data[1]["department"])
	}
}

func TestLoadCSV_HeaderOnly(t *testing.T) {
	data, err := LoadCSV(strings.NewReader("username,department\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Errorf("expected 0 rows, got %d", len(data))
	}
}

func TestLoadCSV_Empty(t *testing.T) {
	data, err := LoadCSV(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if data != nil {
		t.Errorf("expected nil, got %v", data)
	}
}

func TestLoadJSON_Basic(t *testing.T) {
	input := `[{"username": "jdoe", "count": 42}, {"username": "asmith", "active": true}]`
	data, err := LoadJSON(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(data))
	}
	if data[0]["username"] != "jdoe" {
		t.Errorf("expected jdoe, got %s", data[0]["username"])
	}
	// Numbers are converted to string
	if data[0]["count"] != "42" {
		t.Errorf("expected '42' (string), got %s", data[0]["count"])
	}
	// Booleans are converted to string
	if data[1]["active"] != "true" {
		t.Errorf("expected 'true' (string), got %s", data[1]["active"])
	}
}

func TestLoadJSON_EmptyArray(t *testing.T) {
	data, err := LoadJSON(strings.NewReader("[]"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Errorf("expected 0 rows, got %d", len(data))
	}
}

func TestLoadJSON_InvalidJSON(t *testing.T) {
	_, err := LoadJSON(strings.NewReader(`{invalid`))
	if err == nil {
		t.Fatal("expected error")
	}
}
