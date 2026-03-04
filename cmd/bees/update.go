package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/w-h-a/bees/internal/domain"
)

func newUpdateCmd() *cobra.Command {
	var (
		title    string
		desc     string
		status   string
		typ      string
		priority int
		assignee string
		estimate int
		parent   string
		labels   []string
		deferStr string
		dueStr   string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var update domain.IssueUpdate

			if cmd.Flags().Changed("title") {
				update.Title = &title
			}

			if cmd.Flags().Changed("desc") {
				update.Description = &desc
			}

			if cmd.Flags().Changed("status") {
				s := domain.Status(status)
				switch s {
				case domain.StatusOpen, domain.StatusInProgress, domain.StatusApproved, domain.StatusRejected, domain.StatusClosed:
				default:
					return fmt.Errorf("invalid status %q: must be one of open, in_progress, approved, rejected, closed", status)
				}
				update.Status = &s
			}

			if cmd.Flags().Changed("type") {
				t := domain.Type(typ)
				switch t {
				case domain.TypeTask, domain.TypeBug, domain.TypeFeature, domain.TypeChore, domain.TypeDecision, domain.TypeEpic:
				default:
					return fmt.Errorf("invalid type %q: must be one of task, bug, feature, chore, decision, epic", typ)
				}
				update.Type = &t
			}

			if cmd.Flags().Changed("priority") {
				update.Priority = &priority
			}

			if cmd.Flags().Changed("assignee") {
				update.Assignee = &assignee
			}

			if cmd.Flags().Changed("estimate") {
				update.EstimateMins = &estimate
			}

			if cmd.Flags().Changed("parent") {
				update.ParentID = &parent
			}

			if cmd.Flags().Changed("label") {
				update.Labels = &labels
			}

			if cmd.Flags().Changed("defer") {
				t, err := time.Parse("2006-01-02", deferStr)
				if err != nil {
					return fmt.Errorf("failed to parse --defer: %w", err)
				}
				update.DeferUntil = &t
			}

			if cmd.Flags().Changed("due") {
				t, err := time.Parse("2006-01-02", dueStr)
				if err != nil {
					return fmt.Errorf("failed to parse --due: %w", err)
				}
				update.DueAt = &t
			}

			issue, err := svc.UpdateIssue(cmd.Context(), args[0], update)
			if err != nil {
				return err
			}

			if !jsonOutput {
				fmt.Printf("Updated %s: %s\n", issue.ID, issue.Title)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(issue)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Title")
	cmd.Flags().StringVar(&desc, "desc", "", "Description")
	cmd.Flags().StringVar(&status, "status", "", "Status (open, in_progress, approved, rejected, closed)")
	cmd.Flags().StringVar(&typ, "type", "", "Type (task, bug, feature, chore, decision, epic)")
	cmd.Flags().IntVar(&priority, "priority", 2, "Priority 0-4")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee")
	cmd.Flags().IntVar(&estimate, "estimate", 0, "Estimate in minutes")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent issue ID")
	cmd.Flags().StringSliceVar(&labels, "label", nil, "Labels (replaces all, repeatable)")
	cmd.Flags().StringVar(&deferStr, "defer", "", "Defer until date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dueStr, "due", "", "Due date (YYYY-MM-DD)")

	return cmd
}
