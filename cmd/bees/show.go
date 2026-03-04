package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show issue details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issue, err := svc.GetIssue(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if !jsonOutput {
				printIssue(issue)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(issue)
		},
	}

	return cmd
}
