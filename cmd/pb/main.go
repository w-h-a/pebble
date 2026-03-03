package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/w-h-a/pebble/internal/client/repo"
	"github.com/w-h-a/pebble/internal/client/repo/sqlite"
	"github.com/w-h-a/pebble/internal/service"
)

var (
	jsonOutput bool
	verbose    bool
	svc        *service.Service
	dbCloser   func() error
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pb",
		Short: "A minimal task tracker for developers who pair with agentic navigators.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose || os.Getenv("PB_DEBUG") == "1" {
				opts := &slog.HandlerOptions{
					Level: slog.LevelDebug,
				}
				var handler slog.Handler
				if jsonOutput {
					handler = slog.NewJSONHandler(os.Stderr, opts)
				} else {
					handler = slog.NewTextHandler(os.Stderr, opts)
				}
				slog.SetDefault(slog.New(handler))
			}

			needsDB := map[string]bool{"create": true, "show": true}
			if !needsDB[cmd.Name()] {
				return nil
			}

			pebbleDir, err := discoverPebbleDir()
			if err != nil {
				return fmt.Errorf("not a pebble project (run pb init)")
			}

			dbPath := filepath.Join(pebbleDir, "pebble.db")
			// TODO: take user configuration as input if want something other than sqlite
			r, err := sqlite.NewRepo(repo.WithLocation(dbPath))
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			dbCloser = r.Close

			prefix, err := readPrefix(pebbleDir)
			if err != nil {
				r.Close()
				return fmt.Errorf("failed to read config: %w", err)
			}

			svc = service.NewService(r, prefix)

			slog.Debug("project discovered", "dir", pebbleDir, "prefix", prefix)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if dbCloser == nil {
				return nil
			}
			return dbCloser()
		},
	}

	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable debug logging")

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newShowCmd())

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
