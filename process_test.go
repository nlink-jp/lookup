package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestEnrichObject_ExactMatch(t *testing.T) {
	data := LookupData{{"username": "jdoe", "department": "Sales", "role": "Manager"}}
	matcher := &Matcher{LookupField: "username", Method: "exact", CaseSensitive: false}
	mapping := &Mapping{InputField: "user", OutputMap: map[string]string{"department": "dept", "role": "role"}}

	obj := map[string]interface{}{"user": "JDOE", "event": "login"}
	result := enrichObject(obj, mapping, data, matcher, false, nil)

	if result["dept"] != "Sales" {
		t.Errorf("expected dept=Sales, got %v", result["dept"])
	}
	if result["role"] != "Manager" {
		t.Errorf("expected role=Manager, got %v", result["role"])
	}
}

func TestEnrichObject_NoOutputClause(t *testing.T) {
	data := LookupData{{"username": "jdoe", "department": "Sales"}}
	matcher := &Matcher{LookupField: "username", Method: "exact"}
	mapping := &Mapping{InputField: "user", OutputMap: map[string]string{}}

	obj := map[string]interface{}{"user": "jdoe"}
	result := enrichObject(obj, mapping, data, matcher, false, nil)

	if result["department"] != "Sales" {
		t.Errorf("expected all fields added, got %v", result)
	}
}

func TestEnrichObject_MissingField(t *testing.T) {
	data := LookupData{{"username": "jdoe"}}
	matcher := &Matcher{LookupField: "username", Method: "exact"}
	mapping := &Mapping{InputField: "user", OutputMap: map[string]string{}}

	obj := map[string]interface{}{"other": "value"}
	result := enrichObject(obj, mapping, data, matcher, false, nil)

	if _, ok := result["username"]; ok {
		t.Error("should not enrich when input field missing")
	}
}

func TestEnrichObject_NonStringField(t *testing.T) {
	data := LookupData{{"username": "jdoe"}}
	matcher := &Matcher{LookupField: "username", Method: "exact"}
	mapping := &Mapping{InputField: "user", OutputMap: map[string]string{}}

	obj := map[string]interface{}{"user": 12345}
	result := enrichObject(obj, mapping, data, matcher, false, nil)

	if _, ok := result["username"]; ok {
		t.Error("should not enrich when field is non-string")
	}
}

func TestEnrichObject_NoMatch(t *testing.T) {
	data := LookupData{{"username": "jdoe"}}
	matcher := &Matcher{LookupField: "username", Method: "exact"}
	mapping := &Mapping{InputField: "user", OutputMap: map[string]string{}}

	obj := map[string]interface{}{"user": "unknown"}
	result := enrichObject(obj, mapping, data, matcher, false, nil)

	if _, ok := result["username"]; ok {
		t.Error("should not enrich when no match")
	}
}

func TestEnrichObject_DNSMode(t *testing.T) {
	resolver := &mockResolver{addrs: []string{"host.example.com."}}
	mapping := &Mapping{InputField: "ip", OutputMap: map[string]string{}}

	obj := map[string]interface{}{"ip": "192.168.1.1"}
	result := enrichObject(obj, mapping, nil, nil, true, resolver)

	if result["hostname"] != "host.example.com" {
		t.Errorf("expected hostname=host.example.com, got %v", result["hostname"])
	}
}

func TestProcessStream_JSONL(t *testing.T) {
	input := `{"user":"jdoe","event":"login"}
{"user":"asmith","event":"logout"}
`
	var buf bytes.Buffer
	fn := func(obj map[string]interface{}) map[string]interface{} {
		obj["enriched"] = "yes"
		return obj
	}
	if err := processStream(&buf, strings.NewReader(input), fn); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	var obj map[string]interface{}
	json.Unmarshal([]byte(lines[0]), &obj)
	if obj["enriched"] != "yes" {
		t.Errorf("expected enriched=yes, got %v", obj["enriched"])
	}
}

func TestProcessStream_Array(t *testing.T) {
	input := `[{"user":"jdoe"},{"user":"asmith"}]`
	var buf bytes.Buffer
	fn := func(obj map[string]interface{}) map[string]interface{} {
		obj["enriched"] = "yes"
		return obj
	}
	if err := processStream(&buf, strings.NewReader(input), fn); err != nil {
		t.Fatal(err)
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &arr); err != nil {
		t.Fatal(err)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(arr))
	}
	if arr[0]["enriched"] != "yes" {
		t.Errorf("expected enriched=yes, got %v", arr[0]["enriched"])
	}
}

func TestProcessStream_EmptyInput(t *testing.T) {
	var buf bytes.Buffer
	if err := processStream(&buf, strings.NewReader(""), func(obj map[string]interface{}) map[string]interface{} { return obj }); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestProcessStream_MalformedJSONL(t *testing.T) {
	input := `{"user":"jdoe"}
not valid json
{"user":"asmith"}
`
	var buf bytes.Buffer
	fn := func(obj map[string]interface{}) map[string]interface{} { return obj }
	if err := processStream(&buf, strings.NewReader(input), fn); err != nil {
		t.Fatal(err)
	}
	// Should skip bad line, process the others
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines (bad line skipped), got %d", len(lines))
	}
}
