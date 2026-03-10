package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/w-h-a/bees/internal/domain"
)

func newExportCmd() *cobra.Command {
	var (
		status   string
		typ      string
		assignee string
		label    string
		output   string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export issues to JSONL",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := domain.ExportFilter{
				Status:   status,
				Type:     typ,
				Assignee: assignee,
				Label:    label,
			}

			w := os.Stdout

			if output != "" {
				f, err := os.Create(output)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer f.Close()
				w = f
			}

			if err := svc.ExportIssues(cmd.Context(), w, filter); err != nil {
				return err
			}

			if output != "" {
				fmt.Printf("Exported to %s\n", output)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "all", "Filter by status (open, in_progress, closed, all)")
	cmd.Flags().StringVar(&typ, "type", "", "Filter by type")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee")
	cmd.Flags().StringVar(&label, "label", "", "Filter by label")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default stdout)")

	return cmd
}
