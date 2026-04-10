package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the fields that can be specified in a YAML config file.
// It mirrors the CLI flags -p (path) and -x (command).
type Config struct {
	Path    string `yaml:"path"`
	Command string `yaml:"command"`
}

// parseYamlFile reads the YAML config file at meta.config, unmarshals it into
// a Config, and writes the resulting path and command back into meta. Fields
// already set on meta (via CLI flags) are not overwritten when the config file
// value is empty. It is a no-op when meta.config is empty. It calls log.Fatal
// if the file cannot be read or parsed.
func parseYamlFile(meta *watchMeta) {
	// if config file path empty, return
	if meta.config == "" {
		return
	}

	// set the global variables
	data, err := os.ReadFile(meta.config)
	if err != nil {
		log.Fatalf("Failed to read the config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse YAML file: %v, error: %v", meta.config, err)
	}

	// update the meta
	if config.Path != "" {
		meta.path = config.Path
	}
	if config.Command != "" {
		meta.cmd = config.Command
	}

	fmt.Print("PATH=")
	fmt.Println(config.Path)
	fmt.Print("COMMAND=")
	fmt.Println(config.Command)
}
