package jsonl

import "github.com/w-h-a/bees/internal/domain"

func mapFromIssue(issue domain.Issue) jsonlIssue {
	ji := jsonlIssue{
		ID:               issue.ID,
		Title:            issue.Title,
		Description:      issue.Description,
		Status:           string(issue.Status),
		IssueType:        string(issue.Type),
		Priority:         issue.Priority,
		Assignee:         issue.Assignee,
		EstimatedMinutes: issue.EstimateMins,
		DeferUntil:       issue.DeferUntil,
		CreatedAt:        issue.CreatedAt,
		UpdatedAt:        issue.UpdatedAt,
		ClosedAt:         issue.ClosedAt,
		Labels:           issue.Labels,
	}

	for _, d := range issue.Dependencies {
		ji.Dependencies = append(ji.Dependencies, jsonlDep{
			IssueID:     d.IssueID,
			DependsOnID: d.DependsOnID,
			Type:        "blocks",
			CreatedAt:   d.CreatedAt,
		})
	}

	if issue.ParentID != nil {
		ji.Dependencies = append(ji.Dependencies, jsonlDep{
			IssueID:     issue.ID,
			DependsOnID: *issue.ParentID,
			Type:        "parent-child",
			CreatedAt:   issue.CreatedAt,
		})
	}

	for _, c := range issue.Comments {
		ji.Comments = append(ji.Comments, jsonlComment{
			ID:        c.ID,
			IssueID:   c.IssueID,
			Author:    c.Author,
			Text:      c.Body,
			CreatedAt: c.CreatedAt,
		})
	}

	return ji
}
