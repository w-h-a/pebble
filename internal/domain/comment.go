package domain

import "time"

type Comment struct {
	ID        int
	IssueID   string
	Author    string
	Body      string
	CreatedAt time.Time
}
