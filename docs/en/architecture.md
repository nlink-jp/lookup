# lookup — Architecture

## Purpose

Pipe-friendly CLI tool that enriches JSON data streams by looking up values
against external data sources (CSV/JSON) or DNS. Reads JSON from stdin,
matches a field value against a lookup table using configurable rules, and
outputs enriched JSON to stdout.

## Modes of Operation

### Data Source Lookup (default)

```
stdin (JSON array or JSONL)
  → parse each object
  → extract lookup field value
  → match against data source using configured method
  → merge matched row fields into object
  → output enriched object
```

### DNS Lookup (`--dns`)

```
stdin (JSON array or JSONL)
  → parse each object
  → extract lookup field value
  → detect IP vs hostname
  → perform reverse (PTR) or forward (A) lookup
  → merge result into object
  → output enriched object
```

### Config Generation (`generate-config`)

```
data source file (CSV/JSON/JSONL)
  → extract column names / keys
  → generate config.json template
  → output to stdout
```

## Module Structure (new design)

```
cmd/
  root.go          CLI entry point, flag parsing, mode dispatch
config/
  config.go        Config struct, Load(io.Reader), Validate()
  mapping.go       Mapping rule parser
  path.go          Data source path resolution (~, relative)
match/
  matcher.go       Matcher interface + factory
  exact.go         Exact string matching
  wildcard.go      Glob-pattern matching (filepath.Match)
  regex.go         Regex matching
  cidr.go          CIDR network matching
source/
  loader.go        Data source interface
  csv.go           CSV loader
  json.go          JSON/JSONL loader
dns/
  resolver.go      DNS forward/reverse lookup, custom server
process/
  enricher.go      Core enrichment logic (processObject)
  stream.go        Input format detection, JSONL/array I/O
generate/
  generate.go      generate-config subcommand
main.go            Wires everything, calls cmd/
```

### Dependency Graph

```
main.go
  └── cmd/root.go
        ├── config/          (config + mapping + path)
        ├── source/          (CSV/JSON loading)
        ├── match/           (matching algorithms)
        ├── dns/             (DNS resolution)
        ├── process/         (enrichment + I/O)
        └── generate/        (config generation)
```

No external dependencies. Standard library only.

## Data Flow

### Enrichment Pipeline

```
reader ──► detectFormat()
              │
    ┌─────────┴──────────┐
    ▼                    ▼
  JSONL              JSON Array
    │                    │
    ▼                    ▼
 line-by-line       json.Unmarshal
    │                    │
    └────────┬───────────┘
             ▼
     processObject(obj, mapping, lookupData, matcher, opts)
             │
     extract input field value
             │
      ┌──── DNS mode? ────┐
      ▼                   ▼
  findMatch()      dnsLookup()
      │                   │
      └────────┬──────────┘
               ▼
       apply OutputMap (field selection + renaming)
               │
       merge into original object
               │
               ▼
           writer
```

## Matching Methods

| Method | Lookup Field | Input Value | Algorithm |
|--------|-------------|-------------|-----------|
| `exact` | literal string | literal string | String equality (case-insensitive by default) |
| `wildcard` | glob pattern | literal string | `filepath.Match` |
| `regex` | regex pattern | literal string | `regexp.MatchString` |
| `cidr` | CIDR notation | IP address | `net.IPNet.Contains` |

All methods except CIDR support `case_sensitive` flag (default: false).

## Configuration

### Config File (`-c`)

```json
{
  "data_source": "./users.csv",
  "matchers": [
    {
      "input_field": "user_lookup",
      "lookup_field": "username",
      "method": "exact",
      "case_sensitive": false
    }
  ]
}
```

### Mapping Rule (`-m`)

```
<config_ref> as <input_field> [OUTPUT <src> [as <dst>], ...]
```

- `config_ref`: references a matcher's `input_field`
- `input_field`: field name in input JSON to extract value from
- `OUTPUT`: optional field selection and renaming
- Without OUTPUT: all fields from matched row are added

### Data Source Path Resolution

1. `~/...` → home directory expansion
2. Absolute path → used as-is
3. Relative path → resolved relative to config file directory

## Input/Output Formats

### Input Detection

First non-whitespace byte determines format:
- `[` → JSON Array (entire input parsed as one array)
- Otherwise → JSONL (line-by-line processing)

### Output Format

- JSON Array input → formatted JSON Array output (2-space indent)
- JSONL input → JSONL output (compact, one object per line)

### No-match Behavior

Object returned unchanged — no fields added, no error.

## DNS Mode

| Input Value | Lookup Type | Output Field |
|-------------|-------------|-------------|
| Valid IP | Reverse (PTR) | `hostname` |
| Not an IP | Forward (A) | `ip` |

Custom server: `--dns-server 8.8.8.8` (port 53 appended if missing).

## Error Semantics

| Condition | Behavior |
|-----------|----------|
| Missing `-m` flag | Fatal exit |
| Missing `-c` flag (non-DNS) | Fatal exit |
| Config file unreadable | Fatal exit |
| Matcher not found | Fatal exit |
| Data source unreadable | Fatal exit |
| Malformed JSONL line | Warning, skip line |
| No match found | Passthrough unchanged |
| Missing field in input | Passthrough unchanged |
| Non-string field value | Passthrough unchanged |
| Regex compile error | Warning, no match |
| DNS resolution failure | Silent, no enrichment |

## Testing Strategy

### Unit-testable Modules

| Module | What to Test |
|--------|-------------|
| `config/` | Config parsing, validation, mapping rule parsing, path resolution |
| `match/` | Each matcher in isolation (exact, wildcard, regex, CIDR) |
| `source/` | CSV/JSON/JSONL loading with io.Reader |
| `dns/` | Resolver with mock net.Resolver |
| `process/` | Enrichment logic with injected dependencies |
| `generate/` | Key extraction from data sources |

### Integration Tests

Built binary with testdata files — preserves existing black-box test coverage.

### Coverage Targets

| Module | Target |
|--------|--------|
| config/ | 90%+ |
| match/ | 95%+ |
| source/ | 90%+ |
| dns/ | 70%+ (mock resolver) |
| process/ | 85%+ |
| generate/ | 80%+ |
| **Overall** | **80%+** (up from 2.4%) |
