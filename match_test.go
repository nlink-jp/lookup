package main

import (
	"testing"
)

func TestMatchExact_CaseInsensitive(t *testing.T) {
	if !matchExact("JDOE", "jdoe", false) {
		t.Error("expected case-insensitive match")
	}
}

func TestMatchExact_CaseSensitive(t *testing.T) {
	if matchExact("JDOE", "jdoe", true) {
		t.Error("expected no match with case-sensitive")
	}
}

func TestMatchExact_CaseSensitiveMatch(t *testing.T) {
	if !matchExact("jdoe", "jdoe", true) {
		t.Error("expected exact match")
	}
}

func TestMatchWildcard_Match(t *testing.T) {
	if !matchWildcard("b-jones", "b-*", false) {
		t.Error("expected wildcard match")
	}
}

func TestMatchWildcard_CaseInsensitive(t *testing.T) {
	if !matchWildcard("B-JONES", "b-*", false) {
		t.Error("expected case-insensitive wildcard match")
	}
}

func TestMatchWildcard_NoMatch(t *testing.T) {
	if matchWildcard("a-smith", "b-*", false) {
		t.Error("expected no match")
	}
}

func TestMatchRegex_Match(t *testing.T) {
	if !matchRegex("scanner-01", "^scanner-.*$", false) {
		t.Error("expected regex match")
	}
}

func TestMatchRegex_CaseInsensitive(t *testing.T) {
	if !matchRegex("SCANNER-01", "^scanner-.*$", false) {
		t.Error("expected case-insensitive regex match")
	}
}

func TestMatchRegex_NoMatch(t *testing.T) {
	if matchRegex("printer-01", "^scanner-.*$", false) {
		t.Error("expected no match")
	}
}

func TestMatchRegex_InvalidPattern(t *testing.T) {
	if matchRegex("test", "[invalid", false) {
		t.Error("expected false for invalid regex")
	}
}

func TestMatchCIDR_Match(t *testing.T) {
	if !matchCIDR("10.20.30.40", "10.0.0.0/8") {
		t.Error("expected CIDR match")
	}
}

func TestMatchCIDR_NoMatch(t *testing.T) {
	if matchCIDR("8.8.8.8", "10.0.0.0/8") {
		t.Error("expected no match")
	}
}

func TestMatchCIDR_InvalidIP(t *testing.T) {
	if matchCIDR("not-an-ip", "10.0.0.0/8") {
		t.Error("expected false for invalid IP")
	}
}

func TestMatchCIDR_InvalidCIDR(t *testing.T) {
	if matchCIDR("10.0.0.1", "not-cidr") {
		t.Error("expected false for invalid CIDR")
	}
}

func TestMatchCIDR_EmptyCIDR(t *testing.T) {
	if matchCIDR("10.0.0.1", "") {
		t.Error("expected false for empty CIDR")
	}
}

func TestFindMatch_ExactMatch(t *testing.T) {
	data := LookupData{
		{"username": "jdoe", "department": "Sales"},
		{"username": "asmith", "department": "Engineering"},
	}
	matcher := &Matcher{LookupField: "username", Method: "exact", CaseSensitive: false}
	result := FindMatch("JDOE", data, matcher)
	if result == nil {
		t.Fatal("expected match")
	}
	if result["department"] != "Sales" {
		t.Errorf("expected Sales, got %s", result["department"])
	}
}

func TestFindMatch_NoMatch(t *testing.T) {
	data := LookupData{{"username": "jdoe"}}
	matcher := &Matcher{LookupField: "username", Method: "exact"}
	result := FindMatch("unknown", data, matcher)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestFindMatch_MissingLookupField(t *testing.T) {
	data := LookupData{{"other": "value"}}
	matcher := &Matcher{LookupField: "username", Method: "exact"}
	result := FindMatch("test", data, matcher)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestFindMatch_FirstMatchReturned(t *testing.T) {
	data := LookupData{
		{"username": "jdoe", "role": "first"},
		{"username": "jdoe", "role": "second"},
	}
	matcher := &Matcher{LookupField: "username", Method: "exact"}
	result := FindMatch("jdoe", data, matcher)
	if result["role"] != "first" {
		t.Errorf("expected first match, got %s", result["role"])
	}
}
