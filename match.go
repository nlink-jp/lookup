package main

import (
	"log"
	"net"
	"path/filepath"
	"regexp"
	"strings"
)

// FindMatch searches data for a row matching value using matcher's method.
// Returns the matched row or nil if no match.
func FindMatch(value string, data LookupData, matcher *Matcher) map[string]string {
	for _, row := range data {
		lookupValue, ok := row[matcher.LookupField]
		if !ok {
			continue
		}

		if matchRow(value, lookupValue, matcher) {
			return row
		}
	}
	return nil
}

func matchRow(inputValue, lookupValue string, matcher *Matcher) bool {
	switch matcher.Method {
	case "exact":
		return matchExact(inputValue, lookupValue, matcher.CaseSensitive)
	case "wildcard":
		return matchWildcard(inputValue, lookupValue, matcher.CaseSensitive)
	case "regex":
		return matchRegex(inputValue, lookupValue, matcher.CaseSensitive)
	case "cidr":
		return matchCIDR(inputValue, lookupValue)
	default:
		log.Printf("Warning: Unknown match method '%s'", matcher.Method)
		return false
	}
}

func matchExact(input, lookup string, caseSensitive bool) bool {
	if caseSensitive {
		return input == lookup
	}
	return strings.EqualFold(input, lookup)
}

func matchWildcard(input, pattern string, caseSensitive bool) bool {
	if !caseSensitive {
		input = strings.ToLower(input)
		pattern = strings.ToLower(pattern)
	}
	matched, err := filepath.Match(pattern, input)
	if err != nil {
		log.Printf("Warning: Error in wildcard match (pattern: %s): %v", pattern, err)
		return false
	}
	return matched
}

func matchRegex(input, pattern string, caseSensitive bool) bool {
	if !caseSensitive {
		input = strings.ToLower(input)
		pattern = strings.ToLower(pattern)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Printf("Warning: Error compiling regex (pattern: %s): %v", pattern, err)
		return false
	}
	return re.MatchString(input)
}

func matchCIDR(input, cidr string) bool {
	ip := net.ParseIP(input)
	if ip == nil {
		return false
	}
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(ip)
}
