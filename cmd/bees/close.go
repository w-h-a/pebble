package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <id>",
		Short: "Close an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issue, changed, err := svc.CloseIssue(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if !jsonOutput {
				if !changed {
					fmt.Printf("Already closed: %s\n", issue.ID)
				} else {
					fmt.Printf("Closed %s: %s\n", issue.ID, issue.Title)
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
