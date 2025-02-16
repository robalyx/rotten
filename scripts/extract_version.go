package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	EngineVersion string `json:"engineVersion"`
}

func main() {
	configPath := filepath.Join("exports", "official", "export_config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading config:", err)
		os.Exit(1)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing config:", err)
		os.Exit(1)
	}

	fmt.Print(config.EngineVersion)
}
