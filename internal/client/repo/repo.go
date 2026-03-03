package repo

import (
	"context"

	"github.com/w-h-a/pebble/internal/domain"
)

type Repo interface {
	CreateIssue(ctx context.Context, issue *domain.Issue) error
	IssueExists(ctx context.Context, id string) (bool, error)
	AddLabels(ctx context.Context, issueID string, labels []string) error
	GetIssue(ctx context.Context, id string) (*domain.Issue, error)
	GetLabels(ctx context.Context, issueID string) ([]string, error)
	GetDependencies(ctx context.Context, issueID string) ([]domain.Dependency, error)
	GetComments(ctx context.Context, issueID string) ([]domain.Comment, error)
	ResolveID(ctx context.Context, partial string) (string, error)
	Close() error
}
