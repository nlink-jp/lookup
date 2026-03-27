# lookup: A Powerful CLI Lookup Tool

`lookup` is a command-line utility inspired by Splunk's powerful `lookup` command. It enriches JSON data streams by adding fields based on matching values in an external data source (like a CSV or JSON file). It's designed to be a flexible and high-performance tool for data enrichment pipelines.

The tool reads JSON objects (either as a JSON Array or as JSON Lines) from standard input, performs lookups based on sophisticated rules, and outputs the enriched JSON objects to standard output.

---

## Features

-   **Multiple Data Sources**: Use either **CSV** or **JSON** files as your lookup table.
-   **Advanced Matching Methods**:
    -   `exact`: Case-sensitive or insensitive exact string matching.
    -   `wildcard`: Glob-style wildcard matching (e.g., `bot-*`).
    -   `regex`: Powerful matching using regular expressions.
    -   `cidr`: Match IP addresses against CIDR blocks (e.g., `10.0.0.0/8`).
-   **Flexible Configuration**: A central JSON configuration file separates lookup logic from your data, allowing for complex matching rules.
-   **Built-in DNS Lookup**: Perform forward (`A` record) or reverse (`PTR` record) DNS lookups as a native feature.
    -   Optionally specify a custom DNS server for queries.
-   **Flexible Field Mapping**: Intuitive syntax (`<config_ref_field> as <input_field> OUTPUT out1 as new1, ...`) to control which fields are matched and how new fields are named.
-   **Handles Multiple Input Formats**: Automatically detects and processes both **JSON Array** and **JSON Lines (JSONL)** from stdin.
-   **Cross-Platform**: Written in Go, it compiles to a single binary with no external dependencies, running on Linux, macOS, and Windows.

---

## Installation

Pre-compiled binaries for macOS, Windows, and Linux are available on the [Releases](https://github.com/nlink-jp/lookup/releases) page.

---

## Configuration

The configuration file defines where your data is and how to match against it.

### Configuration File (`config.json`)

```json
{
  "data_source": "./path/to/your/data.csv",
  "matchers": [
    {
      "input_field": "field_from_stdin",
      "lookup_field": "column_in_data_source",
      "method": "exact",
      "case_sensitive": false
    },
    {
      "input_field": "another_field_from_stdin",
      "lookup_field": "another_column",
      "method": "regex"
    }
  ]
}
```

-   **`data_source`**: (string) The relative or absolute path to your lookup data file (CSV or JSON).
-   **`matchers`**: (array) A list of objects, where each object defines a specific matching rule.
    -   **`input_field`**: A name for this lookup rule. This is what you refer to in the `-m` flag.
    -   **`lookup_field`**: The column/key name in your `data_source` file to match against.
    -   **`method`**: The matching algorithm to use. Supported values:
        -   `"exact"` (default)
        -   `"wildcard"`
        -   `"regex"`
        -   `"cidr"`
    -   **`case_sensitive`**: (boolean, optional) If `true`, the match will be case-sensitive. Defaults to `false`. This applies to `exact`, `wildcard`, and `regex` methods.

### Configuration Helper (`generate-config`)

To make setup easier, `lookup` provides a helper command to generate a configuration template from your data file. It scans your CSV, JSON, or JSONL file and creates a valid `config.json` structure based on the headers or keys it finds.

```sh
./lookup generate-config -file <path_to_your_data_file>
```

-   **`-file <path>`**: The path to your data source file (e.g., `users.csv` or `data.jsonl`).

**Example:** Given a `users.csv` file with columns `username`, `department`, and `role`:

```sh
./lookup generate-config -file users.csv
```

Output (save as `config.json` and edit as needed):

```json
{
  "data_source": "users.csv",
  "matchers": [
    {
      "input_field": "username",
      "lookup_field": "username",
      "method": "exact",
      "case_sensitive": false
    },
    {
      "input_field": "department",
      "lookup_field": "department",
      "method": "exact",
      "case_sensitive": false
    },
    {
      "input_field": "role",
      "lookup_field": "role",
      "method": "exact",
      "case_sensitive": false
    }
  ]
}
```

---

## Usage

The basic command structure is:

```sh
cat input.json | ./lookup -c <config.json> -m "<mapping_rule>"
```

### Command-Line Flags

| Flag           | Description                                                                                                                              | Required |
| :------------- | :--------------------------------------------------------------------------------------------------------------------------------------- | :------- |
| `-c <path>`    | Path to the JSON configuration file that defines the data source and matching rules.                                                     | Yes      |
| `-m <string>`  | The mapping rule that specifies how to link input data to the lookup table. (See [Mapping Syntax](#mapping-syntax) below).               | Yes      |
| `--dns`        | Enables DNS lookup mode. When used, the `-c` flag is ignored.                                                                            | No       |
| `--dns-server` | (Optional) Specifies a custom DNS server for DNS lookups (e.g., `8.8.8.8` or `1.1.1.1:53`). If not set, the system's default resolver is used. | No       |

### Mapping Syntax

The `-m` flag defines the link between the input stream and the lookup table, and controls the output.

```
"CONFIG_REF_FIELD as INPUT_FIELD [OUTPUT original_name1 as new_name1, original_name2 as new_name2]"
```

-   **`CONFIG_REF_FIELD as INPUT_FIELD`**: (Required)
    -   `CONFIG_REF_FIELD` must match an `input_field` in one of your matchers in `config.json`.
    -   `INPUT_FIELD` is the name of the field in the incoming JSON object to use for the lookup.
-   **`OUTPUT ...`**: (Optional)
    -   Controls which fields from the lookup file are added to the output and allows you to rename them.
    -   If omitted, all columns from the matched row are added with their original names.

### Examples

Let's use the following files for our examples.

**`users.csv`** (Data Source)
```csv
username,department,role,building,ip_range
jdoe,Sales,Manager,A,192.168.1.10
asmith,Engineering,Developer,B,192.168.1.25
b-*,Engineering,QA,B,10.0.0.0/8
^scanner-.*$,IT,Service,A,
```

**`lookup_config.json`** (Configuration)
```json
{
  "data_source": "./users.csv",
  "matchers": [
    {
      "input_field": "user_lookup",
      "lookup_field": "username",
      "method": "exact",
      "case_sensitive": false
    },
    {
      "input_field": "host_lookup",
      "lookup_field": "username",
      "method": "wildcard"
    },
    {
      "input_field": "process_lookup",
      "lookup_field": "username",
      "method": "regex"
    },
    {
      "input_field": "ip_lookup",
      "lookup_field": "ip_range",
      "method": "cidr"
    }
  ]
}
```

**`input.jsonl`** (Input Data)
```json
{"timestamp": "2023-10-28T11:00:00Z", "user": "JDOE", "event": "login"}
{"timestamp": "2023-10-28T11:01:00Z", "hostname": "b-jones", "event": "connect"}
{"timestamp": "2023-10-28T11:02:00Z", "process": "scanner-01", "event": "scan"}
{"timestamp": "2023-10-28T11:03:00Z", "client_ip": "10.20.30.40", "event": "access"}
{"timestamp": "2023-10-28T11:04:00Z", "client_ip": "8.8.8.8", "event": "external_access"}
```

#### Example 1: Case-Insensitive `exact` Match

```sh
cat input.jsonl | ./lookup \
  -c lookup_config.json \
  -m "user_lookup as user OUTPUT department as dept, role"
```

**Output (for the first line):**
```json
{"dept":"Sales","event":"login","role":"Manager","timestamp":"2023-10-28T11:00:00Z","user":"JDOE"}
```

#### Example 2: `cidr` Match

```sh
cat input.jsonl | ./lookup \
  -c lookup_config.json \
  -m "ip_lookup as client_ip"
```

**Output (for the fourth line):**
```json
{"building":"B","client_ip":"10.20.30.40","department":"Engineering","event":"access","ip_range":"10.0.0.0/8","role":"QA","timestamp":"2023-10-28T11:03:00Z","username":"b-*"}
```

#### Example 3: DNS Lookup

**Command (using system resolver):**
```sh
cat input.jsonl | ./lookup \
  --dns \
  -m "dns_reverse_lookup as client_ip OUTPUT hostname as resolved_host"
```

**Command (using a custom DNS server):**
```sh
cat input.jsonl | ./lookup \
  --dns \
  --dns-server "8.8.8.8" \
  -m "dns_reverse_lookup as client_ip OUTPUT hostname as resolved_host"
```

**Output (for the last line, may vary):**
```json
{"client_ip":"8.8.8.8","event":"external_access","resolved_host":"dns.google","timestamp":"2023-10-28T11:04:00Z"}
```

---

## Building

To build from source, you need Go and Make installed.

```sh
# Build for the current platform
make build

# Cross-compile for all platforms (Linux amd64/arm64, macOS amd64/arm64, Windows amd64)
make build-all

# Build all binaries and create .zip archives in dist/
make package
```

The build artifacts will be placed in the `dist/` directory.

---

## License

This project is licensed under the [MIT License](LICENSE).
