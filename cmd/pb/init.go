package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/w-h-a/pebble/internal/client/repo"
	"github.com/w-h-a/pebble/internal/client/repo/sqlite"
	"gopkg.in/yaml.v3"
)

func newInitCmd() *cobra.Command {
	var (
		stealth bool
		prefix  string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new pebble project",
		RunE: func(cmd *cobra.Command, args []string) error {
			pebbleDir := filepath.Join(".", ".pebble")

			if err := os.MkdirAll(pebbleDir, 0o755); err != nil {
				return fmt.Errorf("failed to create .pebble directory: %w", err)
			}

			if prefix == "" {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
				prefix = strings.TrimRight(filepath.Base(wd), "-")
			}

			dbPath := filepath.Join(pebbleDir, "pebble.db")
			r, err := sqlite.NewRepo(repo.WithLocation(dbPath))
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			r.Close()

			slog.Debug("database initialized", "path", dbPath)

			cfg := config{IssuePrefix: prefix}
			data, err := yaml.Marshal(&cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			configPath := filepath.Join(pebbleDir, "config.yaml")
			if err := os.WriteFile(configPath, data, 0o644); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			slog.Debug("config written", "path", configPath, "prefix", prefix)

			if stealth {
				if err := addToGitExclude(); err != nil {
					slog.Debug("stealth mode skipped", "reason", err)
				}
			}

			if !jsonOutput {
				fmt.Printf("Initialized pebble project with prefix %q in %s\n", prefix, pebbleDir)
				return nil
			}

			out := map[string]string{
				"status": "initialized",
				"prefix": prefix,
				"path":   pebbleDir,
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(out)
		},
	}

	cmd.Flags().BoolVar(&stealth, "stealth", false, "Add .pebble/ to .git/info/exclude")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Issue ID prefix (default: directory name)")

	return cmd
}
