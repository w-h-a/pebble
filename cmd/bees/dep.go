package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Manage issue dependencies",
	}

	cmd.AddCommand(newDepAddCmd())
	cmd.AddCommand(newDepRemoveCmd())

	return cmd
}

func newDepAddCmd() *cobra.Command {
	var blockedID string

	cmd := &cobra.Command{
		Use:   "add <blocker-id> --blocks <blocked-id>",
		Short: "Add a blocking dependency",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			blocker, blocked, err := svc.AddDependency(cmd.Context(), args[0], blockedID)
			if err != nil {
				return err
			}

			if !jsonOutput {
				fmt.Printf("%s now blocks %s\n", blocker, blocked)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(map[string]string{
				"blocker_id": blocker,
				"blocked_id": blocked,
				"action":     "added",
			})
		},
	}

	cmd.Flags().StringVar(&blockedID, "blocks", "", "ID of the issue being blocked")
	cmd.MarkFlagRequired("blocks")

	return cmd
}

func newDepRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <blocker-id> <blocked-id>",
		Short: "Remove a blocking dependency",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			blocker, blocked, changed, err := svc.RemoveDependency(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}

			if !jsonOutput {
				if !changed {
					fmt.Printf("No dependency: %s does not block %s\n", blocker, blocked)
				} else {
					fmt.Printf("%s no longer blocks %s\n", blocker, blocked)
				}
				return nil
			}

			action := "removed"
			if !changed {
				action = "none"
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(map[string]string{
				"blocker_id": blocker,
				"blocked_id": blocked,
				"action":     action,
			})
		},
	}

	return cmd
}
