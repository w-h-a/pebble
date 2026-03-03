package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/w-h-a/pebble/internal/domain"
)

var (
	statusColors = map[domain.Status]lipgloss.Style{
		domain.StatusOpen:       lipgloss.NewStyle().Foreground(lipgloss.Color("12")),  // blue
		domain.StatusInProgress: lipgloss.NewStyle().Foreground(lipgloss.Color("11")),  // yellow
		domain.StatusApproved:   lipgloss.NewStyle().Foreground(lipgloss.Color("10")),  // green
		domain.StatusRejected:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")),   // red
		domain.StatusClosed:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")), // gray
	}

	priorityLabels = map[int]string{
		0: "P0 critical",
		1: "P1 high",
		2: "P2 medium",
		3: "P3 low",
		4: "P4 backlog",
	}

	headerStyle  = lipgloss.NewStyle().Bold(true)
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // cyan
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	sectionStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
)

func printIssue(issue *domain.Issue) {
	statusStyle := statusColors[issue.Status]
	statusBadge := statusStyle.Render(string(issue.Status))
	typeBadge := dimStyle.Render(string(issue.Type))

	fmt.Printf("%s %s %s\n", statusBadge, typeBadge, headerStyle.Render(issue.ID))

	fmt.Printf("%s\n\n", headerStyle.Render(issue.Title))

	if issue.Priority != nil {
		label := priorityLabels[*issue.Priority]
		fmt.Printf("  Priority:  %s\n", label)
	}

	if issue.Assignee != "" {
		fmt.Printf("  Assignee:  %s\n", issue.Assignee)
	}

	if issue.EstimateMins > 0 {
		fmt.Printf("  Estimate:  %d min\n", issue.EstimateMins)
	}

	if issue.ParentID != nil {
		fmt.Printf("  Parent:    %s\n", *issue.ParentID)
	}

	if issue.DeferUntil != nil {
		fmt.Printf("  Deferred:  %s\n", issue.DeferUntil.Format("2006-01-02"))
	}

	if issue.DueAt != nil {
		fmt.Printf("  Due:       %s\n", issue.DueAt.Format("2006-01-02"))
	}

	fmt.Printf("  Created:   %s\n", issue.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("  Updated:   %s\n", issue.UpdatedAt.Format("2006-01-02 15:04"))

	if issue.ClosedAt != nil {
		fmt.Printf("  Closed:    %s\n", issue.ClosedAt.Format("2006-01-02 15:04"))
	}

	if len(issue.Labels) > 0 {
		fmt.Println()
		fmt.Printf("%s\n", sectionStyle.Render("Labels"))
		for _, l := range issue.Labels {
			fmt.Printf("  %s\n", labelStyle.Render(l))
		}
	}

	if len(issue.Dependencies) > 0 {
		fmt.Println()
		fmt.Printf("%s\n", sectionStyle.Render("Depends on"))
		for _, d := range issue.Dependencies {
			fmt.Printf("  → %s\n", d.DependsOnID)
		}
	}

	if issue.Description != "" {
		fmt.Println()
		fmt.Printf("%s\n", sectionStyle.Render("Description"))
		fmt.Printf("%s\n", issue.Description)
	}

	if len(issue.Comments) > 0 {
		fmt.Println()
		fmt.Printf("%s\n", sectionStyle.Render("Comments"))
		for _, c := range issue.Comments {
			author := c.Author
			if author == "" {
				author = "anonymous"
			}
			ts := dimStyle.Render(c.CreatedAt.Format("2006-01-02 15:04"))
			fmt.Printf("  %s %s\n", ts, headerStyle.Render(author))

			for _, line := range strings.Split(c.Body, "\n") {
				fmt.Printf("    %s\n", line)
			}
			fmt.Println()
		}
	}
}
