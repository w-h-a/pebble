package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newReopenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reopen <id>",
		Short: "Reopen a closed issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issue, changed, err := svc.ReopenIssue(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if !jsonOutput {
				if !changed {
					fmt.Printf("Already open: %s\n", issue.ID)
				} else {
					fmt.Printf("Reopened %s: %s\n", issue.ID, issue.Title)
				}
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(issue)
		},
	}

	return cmd
}
