# lookup

JSON data enrichment CLI tool — enriches JSON streams via CSV/JSON lookup tables or DNS.

- **Language**: Go 1.25+ (stdlib only, no external dependencies)
- **Series**: util-series
- **CLI**: `lookup -c <config.json> -m "<mapping>" < input.jsonl`
- **CLI**: `lookup --dns -m "<mapping>" < input.jsonl`
- **CLI**: `lookup generate-config -file <data.csv> > config.json`
- **Input**: JSON array or JSONL (auto-detected)
- **Output**: Enriched JSON (same format as input)
- **Matching**: exact, wildcard, regex, cidr (case_sensitive configurable)
- **Build**: `make build` → `dist/lookup`
- **Test**: `go test -race -cover ./...`
- **Docs**: `docs/en/` (English), `docs/ja/` (Japanese)
- **Module layout**: Flat (config.go, match.go, source.go, dns.go, process.go, generate.go, path.go)
