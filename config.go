package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type config struct {
	AlwaysExclude []string `json:"alwaysExclude"`
}

func defaultConfig() config {
	return config{
		AlwaysExclude: []string{".git"},
	}
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "trueblocks", "differ", "config.json")
}

func loadConfig() config {
	cfg := defaultConfig()

	path := configPath()
	if path == "" {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var fileCfg config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return cfg
	}

	if len(fileCfg.AlwaysExclude) > 0 {
		cfg.AlwaysExclude = fileCfg.AlwaysExclude
	}

	return cfg
}
