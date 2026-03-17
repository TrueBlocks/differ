package main

import (
	"os"
	"path/filepath"

	appkit "github.com/TrueBlocks/trueblocks-art/packages/appkit/v2"
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

	fileCfg, err := appkit.LoadJSON(path, config{})
	if err != nil {
		return cfg
	}

	if len(fileCfg.AlwaysExclude) > 0 {
		cfg.AlwaysExclude = fileCfg.AlwaysExclude
	}

	return cfg
}
