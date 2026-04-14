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

## Module Structure

```
main.go              CLI flags, execute(), mode dispatch
config.go            Config/Matcher/Mapping structs, LoadConfig(), ParseMapping()
match.go             FindMatch(), matchExact/Wildcard/Regex/CIDR
source.go            LookupData type, LoadCSV(), LoadJSON(), LoadSource()
dns.go               dnsResolver interface, dnsLookup(), newResolver()
process.go           enrichObject(), processStream() (JSONL/Array auto-detect)
generate.go          generateConfig(), extractCSVHeaders(), extractJSONKeys()
path.go              ResolveDataSourcePath()
```

### Dependency Graph

```
main.go (CLI shell)
  └── execute()
        ├── config.go   (LoadConfig, ParseMapping, FindMatcher)
        ├── path.go     (ResolveDataSourcePath)
        ├── source.go   (LoadSource)
        ├── match.go    (FindMatch)
        ├── dns.go      (dnsLookup, newResolver)
        ├── process.go  (enrichObject, processStream)
        └── generate.go (generateConfig)
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
     enrichObject(obj, mapping, data, matcher, dnsMode, resolver)
             │
     extract input field value
             │
      ┌──── DNS mode? ────┐
      ▼                   ▼
  FindMatch()      dnsLookup()
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

### Unit Tests

| Test File | Coverage |
|-----------|----------|
| `config_test.go` | Config parsing, FindMatcher, ParseMapping (100%) |
| `match_test.go` | exact/wildcard/regex/cidr, case sensitivity, edge cases (100% core) |
| `source_test.go` | CSV/JSON loading via io.Reader (93%+) |
| `dns_test.go` | Mock resolver: forward/reverse, success/failure (100%) |
| `process_test.go` | enrichObject, processStream JSONL/Array, malformed input |
| `generate_test.go` | CSV/JSON key extraction, error cases (90%+) |
| `path_test.go` | Path resolution: tilde, absolute, relative (100%) |

### Regression Tests

`main_test.go` runs `execute()` against existing `testdata/` files:
- exact match with OUTPUT field mapping
- wildcard match (all fields)
- regex match
- CIDR match (JSON array input/output)

### Coverage

| Metric | Value |
|--------|-------|
| Overall | 77%+ |
| Core logic (config, match, dns, path) | 95%+ |
| main() / handleGenerateConfigCmd() | 0% (thin CLI shell) |
