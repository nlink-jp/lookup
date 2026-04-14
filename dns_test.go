package main

import (
	"context"
	"errors"
	"testing"
)

// mockResolver is a test double for dnsResolver.
type mockResolver struct {
	addrs []string
	hosts []string
	err   error
}

func (m *mockResolver) LookupAddr(_ context.Context, _ string) ([]string, error) {
	return m.addrs, m.err
}

func (m *mockResolver) LookupHost(_ context.Context, _ string) ([]string, error) {
	return m.hosts, m.err
}

func TestDNSLookup_ReverseSuccess(t *testing.T) {
	resolver := &mockResolver{addrs: []string{"host.example.com."}}
	result := dnsLookup("192.168.1.1", resolver)
	if result == nil {
		t.Fatal("expected result")
	}
	if result["hostname"] != "host.example.com" {
		t.Errorf("expected host.example.com (trimmed dot), got %s", result["hostname"])
	}
}

func TestDNSLookup_ForwardSuccess(t *testing.T) {
	resolver := &mockResolver{hosts: []string{"10.0.0.1"}}
	result := dnsLookup("example.com", resolver)
	if result == nil {
		t.Fatal("expected result")
	}
	if result["ip"] != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", result["ip"])
	}
}

func TestDNSLookup_ReverseFailure(t *testing.T) {
	resolver := &mockResolver{err: errors.New("not found")}
	result := dnsLookup("192.168.1.1", resolver)
	if result != nil {
		t.Errorf("expected nil on error, got %v", result)
	}
}

func TestDNSLookup_ForwardFailure(t *testing.T) {
	resolver := &mockResolver{err: errors.New("not found")}
	result := dnsLookup("example.com", resolver)
	if result != nil {
		t.Errorf("expected nil on error, got %v", result)
	}
}

func TestDNSLookup_ReverseEmpty(t *testing.T) {
	resolver := &mockResolver{addrs: []string{}}
	result := dnsLookup("192.168.1.1", resolver)
	if result != nil {
		t.Errorf("expected nil for empty result, got %v", result)
	}
}

func TestNewResolver_SystemDefault(t *testing.T) {
	r := newResolver("")
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
}

func TestNewResolver_CustomServer(t *testing.T) {
	r := newResolver("8.8.8.8")
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
}

func TestNewResolver_CustomServerWithPort(t *testing.T) {
	r := newResolver("8.8.8.8:53")
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
}
