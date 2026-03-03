package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/w-h-a/pebble/internal/client/repo"
	"github.com/w-h-a/pebble/internal/domain"
	"github.com/w-h-a/pebble/internal/util/idgen"
)

const (
	maxCollisionRetries = 5
)

func (s *Service) CreateIssue(ctx context.Context, issue *domain.Issue) (string, error) {
	issue.SetDefaults()

	var id string

	for nonce := 0; nonce < maxCollisionRetries; nonce++ {
		id = idgen.Generate(s.prefix, issue.Title, issue.Description, nonce)

		exists, err := s.repo.IssueExists(ctx, id)
		if err != nil {
			return "", fmt.Errorf("failed to check for collision: %w", err)
		}

		if !exists {
			break
		}

		slog.Debug("id collision, retrying", "id", id, "nonce", nonce)

		if nonce == maxCollisionRetries-1 {
			return "", fmt.Errorf("failed to generate unique id after %d attempts", maxCollisionRetries)
		}
	}

	issue.ID = id

	if err := s.repo.CreateIssue(ctx, issue); err != nil {
		return "", fmt.Errorf("failed to create issue: %w", err)
	}

	if len(issue.Labels) > 0 {
		if err := s.repo.AddLabels(ctx, id, issue.Labels); err != nil {
			return "", fmt.Errorf("failed to add labels: %w", err)
		}
	}

	slog.Debug("issue created", "id", id)

	return id, nil
}

func (s *Service) GetIssue(ctx context.Context, idOrPrefix string) (*domain.Issue, error) {
	fullID, err := s.repo.ResolveID(ctx, idOrPrefix)
	if err != nil {
		return nil, err
	}

	issue, err := s.repo.GetIssue(ctx, fullID)
	if err != nil {
		return nil, err
	}

	labels, err := s.repo.GetLabels(ctx, fullID)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}
	issue.Labels = labels

	deps, err := s.repo.GetDependencies(ctx, fullID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}
	issue.Dependencies = deps

	comments, err := s.repo.GetComments(ctx, fullID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	issue.Comments = comments

	return issue, nil
}

type Service struct {
	repo   repo.Repo
	prefix string
}

func NewService(repo repo.Repo, prefix string) *Service {
	return &Service{
		repo:   repo,
		prefix: prefix,
	}
}
