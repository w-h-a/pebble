package repo

import (
	"context"
	"time"

	"github.com/w-h-a/pebble/internal/domain"
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
	Close() error
}
