package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <file.jsonl>",
		Short: "Import issues from a JSONL file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer f.Close()

			result, err := svc.ImportIssues(cmd.Context(), f)
			if err != nil {
				return err
			}

			if !jsonOutput {
				fmt.Printf("Imported: %d created, %d updated, %d unchanged", result.Created, result.Updated, result.Unchanged)
				if result.Skipped > 0 {
					fmt.Printf(", %d skipped", result.Skipped)
				}
				fmt.Println()
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(map[string]any{
				"created":   result.Created,
				"updated":   result.Updated,
				"unchanged": result.Unchanged,
				"skipped":   result.Skipped,
			})
		},
	}

	return cmd
}
