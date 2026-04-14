package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDataSourcePath_Absolute(t *testing.T) {
	got := ResolveDataSourcePath("/data/users.csv", "/etc/lookup")
	if got != "/data/users.csv" {
		t.Errorf("expected /data/users.csv, got %s", got)
	}
}

func TestResolveDataSourcePath_Relative(t *testing.T) {
	got := ResolveDataSourcePath("./users.csv", "/etc/lookup")
	if got != filepath.Join("/etc/lookup", "./users.csv") {
		t.Errorf("expected /etc/lookup/users.csv, got %s", got)
	}
}

func TestResolveDataSourcePath_TildeExpansion(t *testing.T) {
	got := ResolveDataSourcePath("~/data/users.csv", "/etc/lookup")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "data/users.csv")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestResolveDataSourcePath_JustFilename(t *testing.T) {
	got := ResolveDataSourcePath("users.csv", "/etc/lookup")
	if got != filepath.Join("/etc/lookup", "users.csv") {
		t.Errorf("expected /etc/lookup/users.csv, got %s", got)
	}
}

func TestResolveDataSourcePath_EmptyConfigDir(t *testing.T) {
	got := ResolveDataSourcePath("users.csv", "")
	if got != "users.csv" {
		t.Errorf("expected users.csv, got %s", got)
	}
}
