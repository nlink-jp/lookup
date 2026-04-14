package main

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolveDataSourcePath resolves a data source path relative to the config
// file directory. Handles ~/ expansion and absolute/relative paths.
func ResolveDataSourcePath(dataSource, configDir string) string {
	if strings.HasPrefix(dataSource, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, dataSource[2:])
		}
	}
	if filepath.IsAbs(dataSource) {
		return dataSource
	}
	return filepath.Join(configDir, dataSource)
}
