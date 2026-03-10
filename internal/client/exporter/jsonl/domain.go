package jsonl

import "time"

type jsonlIssue struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	Status           string         `json:"status"`
	IssueType        string         `json:"issue_type"`
	Priority         *int           `json:"priority"`
	Assignee         string         `json:"assignee"`
	EstimatedMinutes int            `json:"estimated_minutes"`
	DeferUntil       *time.Time     `json:"defer_until"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	ClosedAt         *time.Time     `json:"closed_at"`
	Labels           []string       `json:"labels"`
	Dependencies     []jsonlDep     `json:"dependencies"`
	Comments         []jsonlComment `json:"comments"`
}

type jsonlDep struct {
	IssueID     string    `json:"issue_id"`
	DependsOnID string    `json:"depends_on_id"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
}

type jsonlComment struct {
	ID        int64     `json:"id"`
	IssueID   string    `json:"issue_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
