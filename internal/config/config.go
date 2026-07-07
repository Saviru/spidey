package config

import (
	"os"
	"path/filepath"

	json "github.com/goccy/go-json"
)

type Directories struct {
	PublicDir string `json:"publicDir"`
	OutputDir string `json:"outputDir"`
}

type Config struct {
	Port        int         `json:"port"`
	Directories Directories `json:"directories"`
}

func LoadConfig(projectDir string) *Config {
	// Defaults
	cfg := &Config{
		Port: 3000,
		Directories: Directories{
			PublicDir: "public",
			OutputDir: "bin/server",
		},
	}

	configPath := filepath.Join(projectDir, "spidey.config.json")
	if file, err := os.Open(configPath); err == nil {
		defer file.Close()
		json.NewDecoder(file).Decode(cfg)
	}

	// Basic fallbacks just in case the json is missing some fields but present
	if cfg.Port == 0 {
		cfg.Port = 3000
	}
	if cfg.Directories.PublicDir == "" {
		cfg.Directories.PublicDir = "public"
	}
	if cfg.Directories.OutputDir == "" {
		cfg.Directories.OutputDir = "bin/server"
	}

	return cfg
}
