package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var version = "dev"

type options struct {
	configFile string
	mappingStr string
	dnsMode    bool
	dnsServer  string
}

// execute runs the lookup pipeline with the given options and I/O.
func execute(opts options, stdin io.Reader, stdout io.Writer) error {
	mapping, err := ParseMapping(opts.mappingStr)
	if err != nil {
		return err
	}

	var data LookupData
	var matcher *Matcher
	var resolver dnsResolver

	if opts.dnsMode {
		if opts.configFile != "" {
			log.Println("Warning: -c flag is ignored when --dns is specified.")
		}
		resolver = newResolver(opts.dnsServer)
	} else {
		if opts.configFile == "" {
			return fmt.Errorf("-c (config file) is required unless --dns is specified")
		}
		configReader, err := os.Open(opts.configFile)
		if err != nil {
			return fmt.Errorf("could not open config file: %w", err)
		}
		defer configReader.Close()

		cfg, err := LoadConfig(configReader)
		if err != nil {
			return err
		}

		matcher, err = cfg.FindMatcher(mapping.ConfigRefField)
		if err != nil {
			return err
		}

		configDir := filepath.Dir(opts.configFile)
		dataPath := ResolveDataSourcePath(cfg.DataSource, configDir)
		data, err = LoadSource(dataPath)
		if err != nil {
			return err
		}
	}

	fn := func(obj map[string]interface{}) map[string]interface{} {
		return enrichObject(obj, mapping, data, matcher, opts.dnsMode, resolver)
	}

	return processStream(stdout, stdin, fn)
}

func main() {
	configFile := flag.String("c", "", "Path to the lookup configuration JSON file.")
	mappingStr := flag.String("m", "", "Mapping rule string.")
	isDnsLookup := flag.Bool("dns", false, "Enable DNS lookup mode.")
	dnsServerAddr := flag.String("dns-server", "", "Custom DNS server address.")
	showVersion := flag.Bool("version", false, "Print version and exit.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `lookup: Enrich JSON/JSONL data by looking up values from external data sources.

Usage:
  lookup -c <config.json> -m "<mapping_rule>" < input.jsonl
  lookup --dns -m "<mapping_rule>" < input.jsonl
  lookup generate-config -file <data_source.csv/json> > config.json
  lookup --version

Options:
`)
		flag.PrintDefaults()
	}

	// Check for generate-config subcommand before flag.Parse
	if len(os.Args) > 1 && os.Args[1] == "generate-config" {
		handleGenerateConfigCmd()
		return
	}

	flag.Parse()
	log.SetOutput(os.Stderr)

	if *showVersion {
		fmt.Printf("lookup version %s\n", version)
		os.Exit(0)
	}

	if *mappingStr == "" {
		log.Fatal("Error: -m (mapping) flag is required.")
	}

	opts := options{
		configFile: *configFile,
		mappingStr: *mappingStr,
		dnsMode:    *isDnsLookup,
		dnsServer:  *dnsServerAddr,
	}

	if err := execute(opts, os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func handleGenerateConfigCmd() {
	fs := flag.NewFlagSet("generate-config", flag.ExitOnError)
	filePath := fs.String("file", "", "Path to the data source file (CSV or JSON).")
	fs.Parse(os.Args[2:])

	if *filePath == "" {
		log.Fatal("Error: -file flag is required for generate-config.")
	}

	if err := generateConfig(os.Stdout, *filePath); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
