package repo

import (
	"context"
	"time"

	"github.com/w-h-a/bees/internal/domain"
)

type Repo interface {
	CreateIssue(ctx context.Context, issue *domain.Issue) error
	IssueExists(ctx context.Context, id string) (bool, error)
	GetIssue(ctx context.Context, id string) (*domain.Issue, error)
	GetLabels(ctx context.Context, issueID string) ([]string, error)
	GetDependencies(ctx context.Context, issueID string) ([]domain.Dependency, error)
	GetComments(ctx context.Context, issueID string) ([]domain.Comment, error)
	ResolveID(ctx context.Context, partial string) (string, error)
	ListIssues(ctx context.Context, filter domain.ListFilter) ([]domain.Issue, error)
	UpdateIssue(ctx context.Context, issue *domain.Issue) error
	CloseIssue(ctx context.Context, id string, now time.Time) error
	ReopenIssue(ctx context.Context, id string, now time.Time) error
	ReadyIssues(ctx context.Context, sort string, limit int) ([]domain.Issue, error)
	UpcomingIssues(ctx context.Context, now time.Time, days int, assignee string) ([]domain.Issue, error)
	Close() error
}
