package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func newReadyCmd() *cobra.Command {
	var (
		sort  string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "ready",
		Short: "Show issues ready to work on",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			issues, err := svc.ReadyIssues(cmd.Context(), sort, limit)
			if err != nil {
				return err
			}

			if !jsonOutput {
				printIssueTable(issues)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(issues)
		},
	}

	cmd.Flags().StringVar(&sort, "sort", "", `Sort by field (priority, created, updated) (default "priority")`)
	cmd.Flags().IntVar(&limit, "limit", 0, "Max results (default 20)")

	return cmd
}
