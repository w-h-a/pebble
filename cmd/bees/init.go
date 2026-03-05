package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/w-h-a/bees/internal/client/repo"
	"github.com/w-h-a/bees/internal/client/repo/sqlite"
)

func newInitCmd() *cobra.Command {
	var (
		stealth bool
		prefix  string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new bees project",
		RunE: func(cmd *cobra.Command, args []string) error {
			beesDir := filepath.Join(".", ".bees")

			if err := os.MkdirAll(beesDir, 0o755); err != nil {
				return fmt.Errorf("failed to create .bees directory: %w", err)
			}

			if prefix == "" {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
				prefix = strings.TrimRight(filepath.Base(wd), "-")
			}

			dbPath := filepath.Join(beesDir, "bees.db")
			r, err := sqlite.NewRepo(repo.WithLocation(dbPath))
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			r.Close()

			slog.Debug("database initialized", "path", dbPath)

			cfg := newConfig()
			if err := cfg.Set("issue-prefix", prefix); err != nil {
				return fmt.Errorf("failed to set prefix: %w", err)
			}

			if err := saveConfig(beesDir, cfg); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			slog.Debug("config written", "path", filepath.Join(beesDir, "config.yaml"), "prefix", prefix)

			if stealth {
				if err := addToGitExclude(); err != nil {
					slog.Debug("stealth mode skipped", "reason", err)
				}
			}

			if !jsonOutput {
				fmt.Printf("Initialized bees project with prefix %q in %s\n", prefix, beesDir)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(map[string]string{
				"status": "initialized",
				"prefix": prefix,
				"path":   beesDir,
			})
		},
	}

	cmd.Flags().BoolVar(&stealth, "stealth", false, "Add .bees/ to .git/info/exclude")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Issue ID prefix (default: directory name)")

	return cmd
}
