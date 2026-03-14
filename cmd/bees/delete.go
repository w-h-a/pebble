package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/w-h-a/bees/internal/domain"
	"github.com/w-h-a/bees/internal/util/duration"
)

func newDeleteCmd() *cobra.Command {
	var (
		closedBefore string
		yes          bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete closed issues in bulk",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if closedBefore == "" {
				return fmt.Errorf("--closed-before is required")
			}

			cutoff, err := duration.Parse(closedBefore)
			if err != nil {
				return fmt.Errorf("failed to parse --closed-before value %q: %w", closedBefore, err)
			}

			filter := domain.DeleteFilter{ClosedBefore: cutoff}

			if !yes {
				candidates, err := svc.PreviewDeleteIssues(cmd.Context(), filter)
				if err != nil {
					return err
				}

				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", " ")
					return enc.Encode(candidates)
				}

				for i, c := range candidates {
					if i >= 20 {
						fmt.Printf("... and %d more\n", len(candidates)-20)
						break
					}
					title := c.Title
					if len(title) > 50 {
						title = title[:47] + "..."
					}
					closedAt := ""
					if c.ClosedAt != nil {
						closedAt = c.ClosedAt.Format("2006-01-02")
					}
					fmt.Printf("  %s  %-12s  %s\n", c.ID, closedAt, title)
				}

				fmt.Printf("Would delete %d issues closed before %s. Run with --yes to confirm.\n",
					len(candidates), cutoff.Format("2006-01-02"))

				return nil
			}

			count, err := svc.DeleteIssues(cmd.Context(), filter)
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", " ")
				return enc.Encode(map[string]any{"deleted": count})
			}

			fmt.Printf("Deleted %d issues closed before %s.\n",
				count, cutoff.Format("2006-01-02"))

			return nil
		},
	}

	cmd.Flags().StringVar(&closedBefore, "closed-before", "", "Delete issues closed before duration (e.g. 12mo, 1y)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm deletion (without this flag, only previews)")

	return cmd
}
