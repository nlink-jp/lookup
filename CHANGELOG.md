# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-04-14

### Changed

- **Full reimplementation** — rewritten from scratch for testability and maintainability
- Split monolithic main.go (636 lines) into 8 focused modules: config, match, source,
  dns, process, generate, path, main
- All core logic accepts io.Reader/io.Writer — fully unit-testable
- DNS resolver abstracted behind interface for mock testing
- Test coverage: 2.4% → 77% (52 tests, up from 9)
- Architecture documentation added (docs/en/, docs/ja/)
- Added README.ja.md, AGENTS.md, CLAUDE.md

### Migration notes

- CLI interface is **fully backward-compatible** — no flag changes
- Config file format unchanged
- Mapping rule syntax unchanged
- All 4 matching methods (exact, wildcard, regex, cidr) behave identically
- JSON Array and JSONL auto-detection preserved

## [1.4.3] - 2026-04-14

### Fixed
- Renamed remaining `lookup-go` references in help text, version output, and
  test code to `lookup`

## [1.4.2] - 2026-03-28

### Changed
- Unified Makefile: replaced macOS universal binary with separate `darwin/amd64` and `darwin/arm64` targets; standardized targets (`build`, `build-all`, `test`, `lint`, `check`, `package`, `clean`, `help`) and output layout (`dist/` flat directory, `.zip` archives).

## [1.4.1] - 2026-03-28

### Internal

- Updated Go module path to `github.com/nlink-jp/lookup` and renamed binary from `lookup-go` to `lookup` following repository transfer and rename to nlink-jp organization.

## [1.4.0] - 2025-09-18

### Changed
- **BREAKING CHANGE**: The logic for the `-m` mapping flag has been redesigned to be more intuitive and explicit.
  - The format is now `"<config_ref_field> as <input_field>"`.
  - `<config_ref_field>` refers to the `input_field` of a matcher in your `config.json`.
  - `<input_field>` refers to the field in the incoming JSON data from which to get the lookup value.
  - This separates the reference to the lookup rule from the key of the data being processed, preventing ambiguity.

## [1.3.0] - 2025-09-10

### Added

-   `~` (tilde) in the `data_source` path of the configuration file is now expanded to the user's home directory.

## [1.2.0] - 2025-09-10

### Changed

-   **Improved Help Message**: The command-line help (`--help`) has been significantly enhanced to be more user-friendly. It now includes a detailed description, usage patterns, subcommand explanations, and practical examples to make the tool easier to understand and use.

## [1.1.0] - 2025-09-10

### Added

-   **`generate-config` Subcommand**: A new helper command to automatically generate a configuration file template from a given data source (`.csv`, `.json`, or `.jsonl`). This simplifies the initial setup process.
    -   It intelligently scans the entire data file to find all possible lookup keys.
    -   It automatically populates the `input_field` and `lookup_field` in the generated template.

## [1.0.0] - 2025-08-26

This is the initial public release of `lookup-go`.

### Added

-   **Core Lookup Functionality**: Enrich JSON/JSONL data from stdin by looking up values in external CSV or JSON data sources.
-   **Advanced Matching Methods**: 
    -   `exact`: Case-sensitive or insensitive exact string matching.
    -   `wildcard`: Glob-style wildcard matching.
    -   `regex`: Regular expression matching.
    -   `cidr`: IP address against CIDR block matching.
-   **DNS Lookup Mode**: Perform forward (A) or reverse (PTR) DNS lookups as a native feature, with support for custom DNS servers.
-   **Flexible I/O**: Automatically handles both JSON Array and JSON Lines (JSONL) input formats.
-   **Robust Build System**: A comprehensive `Makefile` for testing, building, cross-compiling, and packaging releases.
-   **Automated Testing**: A full black-box test suite (`make test`) to ensure reliability.
-   **MIT License**: The project is licensed under the MIT License.
-   **Initial Documentation**: A `README.md` file with detailed usage instructions and examples.