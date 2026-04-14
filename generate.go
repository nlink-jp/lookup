package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// generateConfig reads a data source file, extracts keys/headers, and writes
// a config JSON template to w.
func generateConfig(w io.Writer, dataPath string) error {
	ext := strings.ToLower(filepath.Ext(dataPath))

	file, err := os.Open(dataPath)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", dataPath, err)
	}
	defer file.Close()

	var keys []string
	switch ext {
	case ".csv":
		keys, err = extractCSVHeaders(file)
	case ".json", ".jsonl":
		keys, err = extractJSONKeys(file)
	default:
		return fmt.Errorf("unsupported file type: %s (supported: .csv, .json, .jsonl)", ext)
	}
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return fmt.Errorf("no keys/headers found in %s", dataPath)
	}

	matchers := make([]Matcher, 0, len(keys))
	for _, k := range keys {
		matchers = append(matchers, Matcher{
			InputField:    k,
			LookupField:   k,
			Method:        "exact",
			CaseSensitive: false,
		})
	}

	cfg := Config{
		DataSource: dataPath,
		Matchers:   matchers,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

func extractCSVHeaders(r io.Reader) ([]string, error) {
	reader := csv.NewReader(r)
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read CSV headers: %w", err)
	}
	return headers, nil
}

func extractJSONKeys(r io.Reader) ([]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) == 0 {
		return nil, nil
	}

	keySet := make(map[string]bool)

	if trimmed[0] == '[' {
		// JSON array
		var arr []map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
			return nil, fmt.Errorf("could not parse JSON array: %w", err)
		}
		for _, obj := range arr {
			for k := range obj {
				keySet[k] = true
			}
		}
	} else {
		// JSONL
		scanner := bufio.NewScanner(strings.NewReader(trimmed))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				continue // skip invalid lines
			}
			for k := range obj {
				keySet[k] = true
			}
		}
	}

	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	return keys, nil
}
