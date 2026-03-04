package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func newUpcomingCmd() *cobra.Command {
	var (
		days     int
		assignee string
	)

	cmd := &cobra.Command{
		Use:   "upcoming",
		Short: "Show issues scheduled for the coming days",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			issues, err := svc.UpcomingIssues(cmd.Context(), days, assignee)
			if err != nil {
				return err
			}

			if !jsonOutput {
				printUpcomingTable(issues)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(issues)
		},
	}

	cmd.Flags().IntVar(&days, "days", 0, "Lookahead window in days (default 15)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee")

	return cmd
}
