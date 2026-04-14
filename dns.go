package main

import (
	"context"
	"net"
	"strings"
	"time"
)

// dnsResolver abstracts DNS resolution for testability.
type dnsResolver interface {
	LookupAddr(ctx context.Context, addr string) ([]string, error)
	LookupHost(ctx context.Context, host string) ([]string, error)
}

// newResolver creates a DNS resolver. If server is non-empty, uses a custom
// DNS server; otherwise uses the system resolver.
func newResolver(server string) dnsResolver {
	if server == "" {
		return net.DefaultResolver
	}
	if !strings.Contains(server, ":") {
		server += ":53"
	}
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", server)
		},
	}
}

// dnsLookup performs a forward or reverse DNS lookup based on whether value
// is an IP address. Returns a map with "hostname" or "ip" key, or nil on failure.
func dnsLookup(value string, resolver dnsResolver) map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := net.ParseIP(value)
	if ip != nil {
		// Reverse lookup: IP → hostname
		names, err := resolver.LookupAddr(ctx, value)
		if err != nil || len(names) == 0 {
			return nil
		}
		hostname := strings.TrimSuffix(names[0], ".")
		return map[string]string{"hostname": hostname}
	}

	// Forward lookup: hostname → IP
	addrs, err := resolver.LookupHost(ctx, value)
	if err != nil || len(addrs) == 0 {
		return nil
	}
	return map[string]string{"ip": addrs[0]}
}
