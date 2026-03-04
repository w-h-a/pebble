package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/w-h-a/bees/internal/client/repo"
	"github.com/w-h-a/bees/internal/client/repo/sqlite"
	"github.com/w-h-a/bees/internal/service"
)

var (
	jsonOutput bool
	verbose    bool
	svc        *service.Service
	dbCloser   func() error
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bees",
		Short: "A minimal task tracker for developers who pair with agentic navigators.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose || os.Getenv("BEES_DEBUG") == "1" {
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

			needsDB := map[string]bool{
				"create":   true,
				"show":     true,
				"list":     true,
				"update":   true,
				"close":    true,
				"reopen":   true,
				"ready":    true,
				"upcoming": true,
			}
			if !needsDB[cmd.Name()] {
				return nil
			}

			beesDir, err := discoverBeesDir()
			if err != nil {
				return fmt.Errorf("not a bees project (run bees init)")
			}

			dbPath := filepath.Join(beesDir, "bees.db")
			r, err := sqlite.NewRepo(repo.WithLocation(dbPath))
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			dbCloser = r.Close

			prefix, err := readPrefix(beesDir)
			if err != nil {
				r.Close()
				return fmt.Errorf("failed to read config: %w", err)
			}

			svc = service.NewService(r, prefix)

			slog.Debug("project discovered", "dir", beesDir, "prefix", prefix)

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
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newCloseCmd())
	cmd.AddCommand(newReopenCmd())
	cmd.AddCommand(newReadyCmd())
	cmd.AddCommand(newUpcomingCmd())

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
