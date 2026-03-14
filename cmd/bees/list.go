package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/w-h-a/bees/internal/domain"
	"github.com/w-h-a/bees/internal/util/duration"
)

func newListCmd() *cobra.Command {
	var (
		status   string
		typ      string
		assignee string
		label    string
		parent   string
		sort     string
		limit    int
		since    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := domain.ListFilter{
				Status:   status,
				Type:     typ,
				Assignee: assignee,
				Label:    label,
				Parent:   parent,
				Sort:     sort,
				Limit:    limit,
			}

			if since != "" {
				t, err := duration.Parse(since)
				if err != nil {
					return fmt.Errorf("failed to parse --since value %q: %w", since, err)
				}
				filter.Since = &t
			}

			issues, err := svc.ListIssues(cmd.Context(), filter)
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

	cmd.Flags().StringVar(&status, "status", "", `Filter by status (open, in_progress, closed, all) (default "open")`)
	cmd.Flags().StringVar(&typ, "type", "", "Filter by type")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee")
	cmd.Flags().StringVar(&label, "label", "", "Filter by label")
	cmd.Flags().StringVar(&parent, "parent", "", "Filter by parent (shows all descendants)")
	cmd.Flags().StringVar(&sort, "sort", "", `Sort by field (priority, created, updated) (default "priority")`)
	cmd.Flags().IntVar(&limit, "limit", 0, "Max results (default 50)")
	cmd.Flags().StringVar(&since, "since", "", `Show issues closed since duration (e.g. 6mo, 1y, 30d) — requires --status closed`)

	return cmd
}
