package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	noopexporter "github.com/w-h-a/bees/internal/client/exporter/noop"
	"github.com/w-h-a/bees/internal/client/importer/beads"
	noopimporter "github.com/w-h-a/bees/internal/client/importer/noop"
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
	var (
		beesDir string
		cfg     *config
	)

	cmd := &cobra.Command{
		Use:   "bees",
		Short: "An alternative to a sea of .md files for developers who pair with agentic navigators.",
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

			if cmd.Name() == "init" {
				return nil
			}

			var err error

			beesDir, err = discoverBeesDir()
			if err != nil {
				return fmt.Errorf("not a bees project (run bees init)")
			}

			cfg, err = loadConfig(beesDir)
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}

			if !cmd.Flags().Changed("json") {
				val, _ := resolveConfig(cfg, "json")
				if val == "true" || val == "1" {
					jsonOutput = true
				}
			}

			prefix := cfg.IssuePrefix()

			slog.Debug("project discovered", "dir", beesDir, "prefix", prefix)

			slog.Debug("command path", "path", cmd.CommandPath(), "name", cmd.Name())

			needsDB := map[string]bool{
				"bees create":     true,
				"bees show":       true,
				"bees list":       true,
				"bees search":     true,
				"bees update":     true,
				"bees close":      true,
				"bees reopen":     true,
				"bees ready":      true,
				"bees upcoming":   true,
				"bees dep add":    true,
				"bees dep remove": true,
				"bees dep graph":  true,
				"bees comment":    true,
				"bees import":     true,
			}
			if !needsDB[cmd.CommandPath()] {
				return nil
			}

			dbPath := filepath.Join(beesDir, "bees.db")
			r, err := sqlite.NewRepo(repo.WithLocation(dbPath))
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			dbCloser = r.Close

			i, _ := noopimporter.NewImporter()
			if cmd.CommandPath() == "bees import" {
				i, err = beads.NewImporter()
				if err != nil {
					return fmt.Errorf("failed to initialize importer: %w", err)
				}
			}

			e, _ := noopexporter.NewExporter()

			svc = service.NewService(r, i, e, prefix)

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
	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newCloseCmd())
	cmd.AddCommand(newReopenCmd())
	cmd.AddCommand(newReadyCmd())
	cmd.AddCommand(newUpcomingCmd())
	cmd.AddCommand(newDepCmd())
	cmd.AddCommand(newCommentCmd(&cfg))
	cmd.AddCommand(newConfigCmd(&beesDir, &cfg))
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
