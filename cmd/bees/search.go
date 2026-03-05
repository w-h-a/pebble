package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search issues by title or description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issues, err := svc.SearchIssues(cmd.Context(), args[0], limit)
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

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (default is 50)")

	return cmd
}
