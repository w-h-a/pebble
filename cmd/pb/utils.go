package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func discoverPebbleDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, ".pebble")
		if info, err := os.Stat(filepath.Join(candidate, "pebble.db")); err == nil && !info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .pebble directory found")
		}

		dir = parent
	}
}

type config struct {
	IssuePrefix string `yaml:"issue-prefix"`
}

func readPrefix(pebbleDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(pebbleDir, "config.yaml"))
	if err != nil {
		return "", fmt.Errorf("failed to read config.yaml: %w", err)
	}

	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	return cfg.IssuePrefix, nil
}

func addToGitExclude() error {
	excludePath := filepath.Join(".git", "info", "exclude")

	if err := os.MkdirAll(filepath.Dir(excludePath), 0o755); err != nil {
		return fmt.Errorf("failed to create .git/info: %w", err)
	}

	existing, err := os.ReadFile(excludePath)
	if err == nil && strings.Contains(string(existing), ".pebble/") {
		slog.Debug("already in .git/info/exclude")
		return nil
	}

	f, err := os.OpenFile(excludePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open exclude file: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, ".pebble/"); err != nil {
		return fmt.Errorf("failed to write exclude entry: %w", err)
	}

	slog.Debug("added .pebble/ to .git/info/exclude")

	return nil
}
