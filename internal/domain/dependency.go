package domain

import "time"

type Dependency struct {
	IssueID     string
	DependsOnID string
	CreatedAt   time.Time
}
