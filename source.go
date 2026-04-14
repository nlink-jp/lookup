package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LookupData is a slice of rows, each row mapping field names to string values.
type LookupData []map[string]string

// LoadCSV reads CSV data (header + rows) from r.
func LoadCSV(r io.Reader) (LookupData, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read CSV: %w", err)
	}
	if len(records) < 1 {
		return nil, nil
	}

	headers := records[0]
	var data LookupData
	for _, row := range records[1:] {
		entry := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				entry[h] = row[i]
			}
		}
		data = append(data, entry)
	}
	return data, nil
}

// LoadJSON reads a JSON array of objects from r, converting all values to strings.
func LoadJSON(r io.Reader) (LookupData, error) {
	var raw []map[string]interface{}
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}

	data := make(LookupData, 0, len(raw))
	for _, obj := range raw {
		entry := make(map[string]string, len(obj))
		for k, v := range obj {
			entry[k] = fmt.Sprintf("%v", v)
		}
		data = append(data, entry)
	}
	return data, nil
}

// LoadSource loads lookup data from a file, dispatching by extension.
func LoadSource(path string) (LookupData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open data source %s: %w", path, err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return LoadCSV(file)
	case ".json", ".jsonl":
		return LoadJSON(file)
	default:
		return nil, fmt.Errorf("unsupported data source format: %s (supported: .csv, .json)", ext)
	}
}
