# CLAUDE.md — lookup

**Organization rules (mandatory): https://github.com/nlink-jp/.github/blob/main/CONVENTIONS.md**

## Purpose

JSON data enrichment CLI tool. Reads JSON from stdin, looks up values in
external data sources (CSV/JSON) or DNS, and outputs enriched JSON to stdout.

## Build & test

```bash
make build       # Build → dist/lookup
make test        # Run tests with race detector + coverage
go test ./...    # Same without Makefile
```

## Architecture

```
main.go         CLI shell: flag parsing → execute()
config.go       Config/Matcher/Mapping + LoadConfig/ParseMapping
match.go        FindMatch + exact/wildcard/regex/cidr matchers
source.go       LookupData + CSV/JSON loaders
dns.go          dnsResolver interface + dnsLookup (mockable)
process.go      enrichObject + processStream (JSONL/Array auto-detect)
generate.go     generate-config subcommand
path.go         Data source path resolution
```

All core logic accepts io.Reader/io.Writer for testability.
No external dependencies — standard library only.

## Key conventions

- Matching methods: exact, wildcard (filepath.Match), regex, cidr (net.IPNet.Contains)
- case_sensitive defaults to false; ignored for CIDR
- JSON Array input → formatted Array output; JSONL → JSONL
- No match → object returned unchanged (passthrough)
- DNS mode: IP → reverse (PTR), hostname → forward (A)
- Malformed JSONL lines → warning + skip (not fatal)

## Communication Language

All communication between contributors and Claude Code is conducted in **Japanese**.
