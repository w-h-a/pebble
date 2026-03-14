package domain

import (
	"errors"
	"time"

	"github.com/w-h-a/bees/internal/util/dfs"
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusApproved   Status = "approved"
	StatusRejected   Status = "rejected"
	StatusClosed     Status = "closed"
)

type Type string

const (
	TypeTask     Type = "task"
	TypeBug      Type = "bug"
	TypeFeature  Type = "feature"
	TypeChore    Type = "chore"
	TypeEpic     Type = "epic"
	TypeDecision Type = "decision"
)

type Issue struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       Status       `json:"status"`
	Type         Type         `json:"type"`
	Priority     *int         `json:"priority"`
	Assignee     string       `json:"assignee"`
	EstimateMins int          `json:"estimate_mins"`
	DeferUntil   *time.Time   `json:"defer_until"`
	DueAt        *time.Time   `json:"due_at"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	ClosedAt     *time.Time   `json:"closed_at"`
	ParentID     *string      `json:"parent_id"`
	Labels       []string     `json:"labels"`
	Dependencies []Dependency `json:"dependencies"`
	Comments     []Comment    `json:"comments"`
}

func (i *Issue) SetDefaults() {
	now := time.Now()

	if i.Status == "" {
		i.Status = StatusOpen
	}

	if i.Type == "" {
		i.Type = TypeTask
	}

	if i.Priority == nil {
		p := 2
		i.Priority = &p
	}

	if i.CreatedAt.IsZero() {
		i.CreatedAt = now
	}

	if i.UpdatedAt.IsZero() {
		i.UpdatedAt = now
	}
}

type ListFilter struct {
	Status   string
	Type     string
	Assignee string
	Label    string
	Parent   string
	Sort     string
	Limit    int
	Since    *time.Time
}

func (f ListFilter) Validate() error {
	if f.Since != nil && f.Status != string(StatusClosed) {
		return errors.New("filtering by 'since' is only supported when status is 'closed'")
	}
	return nil
}

type DeleteFilter struct {
	ClosedBefore time.Time
}

func (f DeleteFilter) Validate() error {
	if f.ClosedBefore.IsZero() {
		return errors.New("deleting requires a 'closed-before' filter")
	}
	return nil
}

type ExportFilter struct {
	Status   string
	Type     string
	Assignee string
	Label    string
}

type IssueUpdate struct {
	Title        *string
	Description  *string
	Status       *Status
	Type         *Type
	Priority     *int
	Assignee     *string
	EstimateMins *int
	ParentID     *string
	DeferUntil   *time.Time
	DueAt        *time.Time
	Labels       *[]string
}

func Descendants(issues []Issue, rootID string) []Issue {
	adj := map[string][]string{}
	issueMap := map[string]Issue{}

	for _, iss := range issues {
		issueMap[iss.ID] = iss
		if iss.ParentID != nil {
			adj[*iss.ParentID] = append(adj[*iss.ParentID], iss.ID)
		}
	}

	reachable := dfs.Reachable(adj, rootID)

	var result []Issue
	for id := range reachable {
		if id == rootID {
			continue
		}
		if iss, ok := issueMap[id]; ok {
			result = append(result, iss)
		}
	}

	return result
}
