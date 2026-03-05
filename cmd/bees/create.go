package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/w-h-a/bees/internal/domain"
)

func newCreateCmd() *cobra.Command {
	var (
		issueType string
		priority  int
		assignee  string
		estimate  int
		desc      string
		parent    string
		labels    []string
		deferStr  string
		dueStr    string
	)

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if issueType != "" {
				switch domain.Type(issueType) {
				case domain.TypeTask, domain.TypeBug, domain.TypeFeature, domain.TypeChore, domain.TypeDecision, domain.TypeEpic:
				default:
					return fmt.Errorf("invalid type %q: must be one of task, bug, feature, chore, decision, epic", issueType)
				}
			}

			issue := domain.Issue{
				Title:       args[0],
				Description: desc,
				Type:        domain.Type(issueType),
				Assignee:    assignee,
				Labels:      labels,
			}

			if cmd.Flags().Changed("priority") {
				issue.Priority = &priority
			}

			if estimate > 0 {
				issue.EstimateMins = estimate
			}

			if parent != "" {
				issue.ParentID = &parent
			}

			if deferStr != "" {
				t, err := time.Parse("2006-01-02", deferStr)
				if err != nil {
					return fmt.Errorf("failed to parse --defer: %w", err)
				}
				issue.DeferUntil = &t
			}

			if dueStr != "" {
				t, err := time.Parse("2006-01-02", dueStr)
				if err != nil {
					return fmt.Errorf("failed to parse --due: %w", err)
				}
				issue.DueAt = &t
			}

			id, err := svc.CreateIssue(cmd.Context(), &issue)
			if err != nil {
				return err
			}

			if !jsonOutput {
				fmt.Printf("Created %s: %s\n", id, issue.Title)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			out := map[string]any{
				"id":     id,
				"title":  issue.Title,
				"type":   string(issue.Type),
				"status": string(issue.Status),
			}

			if issue.Priority != nil {
				out["priority"] = *issue.Priority
			}

			return enc.Encode(out)
		},
	}

	cmd.Flags().StringVar(&issueType, "type", "", "Issue type (task, bug, feature, chore, decision, epic)")
	cmd.Flags().IntVar(&priority, "priority", 2, "Priority 0-4")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee")
	cmd.Flags().IntVar(&estimate, "estimate", 0, "Estimate in minutes")
	cmd.Flags().StringVar(&desc, "desc", "", "Description")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent issue ID")
	cmd.Flags().StringSliceVar(&labels, "label", nil, "Labels (repeatable)")
	cmd.Flags().StringVar(&deferStr, "defer", "", "Defer until date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dueStr, "due", "", "Due date (YYYY-MM-DD)")

	return cmd
}
