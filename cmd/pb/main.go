package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pb",
		Short: "Pebble -- A minimal task tracker for developers who pair with agentic navigators.",
	}

	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
