package repo

import (
	"context"
	"time"

	"github.com/w-h-a/bees/internal/domain"
)

type Repo interface {
	CreateIssue(ctx context.Context, issue *domain.Issue) error
	IssueExists(ctx context.Context, id string) (bool, error)
	ResolveID(ctx context.Context, partial string) (string, error)
	GetIssue(ctx context.Context, id string) (*domain.Issue, error)
	ListIssues(ctx context.Context, filter domain.ListFilter) ([]domain.Issue, error)
	SearchIssues(ctx context.Context, query string, limit int) ([]domain.Issue, error)
	UpdateIssue(ctx context.Context, issue *domain.Issue) error
	CloseIssue(ctx context.Context, id string, now time.Time) error
	ReopenIssue(ctx context.Context, id string, now time.Time) error
	ReadyIssues(ctx context.Context, sort string, limit int) ([]domain.Issue, error)
	UpcomingIssues(ctx context.Context, now time.Time, days int, assignee string) ([]domain.Issue, error)

	GetLabels(ctx context.Context, issueID string) ([]string, error)

	GetDependencies(ctx context.Context, issueID string) ([]domain.Dependency, error)
	AddDependency(ctx context.Context, dep domain.Dependency) error
	AddDependencyIfAcyclic(ctx context.Context, dep domain.Dependency) error
	GetDependencyGraph(ctx context.Context) ([]domain.Dependency, error)
	RemoveDependency(ctx context.Context, issueID, dependsOnID string) (bool, error)

	GetComments(ctx context.Context, issueID string) ([]domain.Comment, error)
	AddComment(ctx context.Context, comment *domain.Comment) error

	Close() error
}
