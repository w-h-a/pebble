package domain

import "time"

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
	ID           string
	Title        string
	Description  string
	Status       Status
	Type         Type
	Priority     *int
	Assignee     string
	EstimateMins int
	DeferUntil   *time.Time
	DueAt        *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ClosedAt     *time.Time
	ParentID     *string
	Labels       []string
	Dependencies []Dependency
	Comments     []Comment
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
