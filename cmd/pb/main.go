package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	verbose    bool
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pb",
		Short: "Pebble -- A minimal task tracker for developers who pair with agentic navigators.",
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
			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable debug logging")

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
