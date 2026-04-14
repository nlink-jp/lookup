package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

// enrichObject looks up a value from the input object and merges matched fields.
func enrichObject(
	obj map[string]interface{},
	mapping *Mapping,
	data LookupData,
	matcher *Matcher,
	dnsMode bool,
	resolver dnsResolver,
) map[string]interface{} {
	inputValue, ok := obj[mapping.InputField]
	if !ok {
		return obj
	}
	inputStr, ok := inputValue.(string)
	if !ok {
		return obj
	}

	var matched map[string]string
	if dnsMode {
		matched = dnsLookup(inputStr, resolver)
	} else {
		matched = FindMatch(inputStr, data, matcher)
	}
	if matched == nil {
		return obj
	}

	// Apply output mapping
	if len(mapping.OutputMap) > 0 {
		for srcField, dstField := range mapping.OutputMap {
			if val, exists := matched[srcField]; exists {
				obj[dstField] = val
			}
		}
	} else {
		// No OUTPUT clause: add all matched fields
		for k, v := range matched {
			obj[k] = v
		}
	}

	return obj
}

// processStream reads input, detects format (JSON array or JSONL), enriches
// each object using fn, and writes to w.
func processStream(w io.Writer, r io.Reader, fn func(map[string]interface{}) map[string]interface{}) error {
	input, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return nil
	}

	if trimmed[0] == '[' {
		return processArray(w, trimmed, fn)
	}
	return processJSONL(w, trimmed, fn)
}

func processArray(w io.Writer, data []byte, fn func(map[string]interface{}) map[string]interface{}) error {
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		return fmt.Errorf("could not parse JSON array: %w", err)
	}

	for i := range arr {
		arr[i] = fn(arr[i])
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(arr)
}

func processJSONL(w io.Writer, data []byte, fn func(map[string]interface{}) map[string]interface{}) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			log.Printf("Warning: Could not parse line as JSON, skipping: %v", err)
			continue
		}

		enriched := fn(obj)
		out, err := json.Marshal(enriched)
		if err != nil {
			log.Printf("Warning: Could not marshal result to JSON, skipping: %v", err)
			continue
		}
		fmt.Fprintln(w, string(out))
	}
	return scanner.Err()
}
