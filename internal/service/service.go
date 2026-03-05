package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/w-h-a/bees/internal/client/importer"
	"github.com/w-h-a/bees/internal/client/repo"
	"github.com/w-h-a/bees/internal/domain"
	"github.com/w-h-a/bees/internal/util/dfs"
	"github.com/w-h-a/bees/internal/util/hash"
	"github.com/w-h-a/bees/internal/util/idgen"
)

const (
	maxCollisionRetries = 5
)

type Service struct {
	repo   repo.Repo
	imp    importer.Importer
	prefix string
}

func (s *Service) ImportIssues(ctx context.Context, r io.Reader) (ImportResult, error) {
	issues, err := s.imp.Parse(r)
	if err != nil {
		return ImportResult{}, fmt.Errorf("failed to parse import data: %w", err)
	}

	slog.Debug("importing issues", "count", len(issues))

	var result ImportResult

	type pending struct {
		issue   domain.Issue
		created bool
	}
	var processed []pending

	for _, issue := range issues {
		issue.SetDefaults()

		exists, err := s.repo.IssueExists(ctx, issue.ID)
		if err != nil {
			slog.Debug("skipping issue", "id", issue.ID, "error", err)
			result.Skipped++
			continue
		}

		if !exists {
			if err := s.repo.CreateIssue(ctx, &issue); err != nil {
				slog.Debug("skipping issue on create", "id", issue.ID, "error", err)
				result.Skipped++
				continue
			}
			result.Created++
			processed = append(processed, pending{issue: issue, created: true})
			slog.Debug("issue created", "id", issue.ID)
			continue
		}

		existing, err := s.repo.GetIssue(ctx, issue.ID)
		if err != nil {
			slog.Debug("skipping issue on get", "id", issue.ID, "error", err)
			result.Skipped++
			continue
		}

		existingLabels, _ := s.repo.GetLabels(ctx, issue.ID)
		existing.Labels = existingLabels

		if hash.Fields(issueFields(existing)) == hash.Fields(issueFields(&issue)) {
			result.Unchanged++
			slog.Debug("issue unchanged", "id", issue.ID)
			continue
		}

		issue.UpdatedAt = time.Now()

		if err := s.repo.UpdateIssue(ctx, &issue); err != nil {
			slog.Debug("skipping issue on update", "id", issue.ID, "error", err)
			result.Skipped++
			continue
		}

		result.Updated++
		processed = append(processed, pending{issue: issue, created: false})

		slog.Debug("issue updated", "id", issue.ID)
	}

	for _, p := range processed {
		for _, dep := range p.issue.Dependencies {
			if err := s.repo.AddDependency(ctx, dep); err != nil {
				slog.Debug("skipping dependency", "issue", dep.IssueID, "depends_on", dep.DependsOnID, "error", err)
			}
		}
	}

	for _, p := range processed {
		if !p.created {
			continue
		}
		for i := range p.issue.Comments {
			if err := s.repo.AddComment(ctx, &p.issue.Comments[i]); err != nil {
				slog.Debug("skipping comment", "issue", p.issue.Comments[i].IssueID, "error", err)
			}
		}
	}

	slog.Debug("import complete",
		"created", result.Created,
		"updated", result.Updated,
		"unchanged", result.Unchanged,
		"skipped", result.Skipped,
	)

	return result, nil
}

func (s *Service) CreateIssue(ctx context.Context, issue *domain.Issue) (string, error) {
	issue.SetDefaults()

	slog.Debug("creating issue",
		"title", issue.Title,
		"type", string(issue.Type),
		"priority", *issue.Priority,
		"label_count", len(issue.Labels),
	)

	var id string

	for nonce := range maxCollisionRetries {
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

	slog.Debug("issue created", "id", id)

	return id, nil
}

func (s *Service) GetIssue(ctx context.Context, idOrPrefix string) (*domain.Issue, error) {
	slog.Debug("getting issue", "id_or_prefix", idOrPrefix)

	fullID, err := s.repo.ResolveID(ctx, idOrPrefix)
	if err != nil {
		return nil, err
	}

	if fullID != idOrPrefix {
		slog.Debug("prefix resolved", "input", idOrPrefix, "resolved", fullID)
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

	slog.Debug("issue retrieved", "id", fullID,
		"label_count", len(labels),
		"dep_count", len(deps),
		"comment_count", len(comments),
	)

	return issue, nil
}

func (s *Service) ListIssues(ctx context.Context, filter domain.ListFilter) ([]domain.Issue, error) {
	if filter.Status == "" {
		filter.Status = "open"
	}

	if filter.Sort == "" {
		filter.Sort = "priority"
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	slog.Debug("listing issues",
		"status", filter.Status,
		"type", filter.Type,
		"assignee", filter.Assignee,
		"label", filter.Label,
		"sort", filter.Sort,
		"limit", filter.Limit,
	)

	issues, err := s.repo.ListIssues(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	if err := s.populateRelations(ctx, issues); err != nil {
		return nil, err
	}

	slog.Debug("issues listed", "count", len(issues))

	return issues, nil
}

func (s *Service) SearchIssues(ctx context.Context, query string, limit int) ([]domain.Issue, error) {
	if limit <= 0 {
		limit = 50
	}

	slog.Debug("searching issues", "query", query, "limit", limit)

	issues, err := s.repo.SearchIssues(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	if err := s.populateRelations(ctx, issues); err != nil {
		return nil, err
	}

	slog.Debug("search results", "count", len(issues))

	return issues, nil
}

func (s *Service) UpdateIssue(ctx context.Context, idOrPrefix string, update domain.IssueUpdate) (*domain.Issue, error) {
	slog.Debug("updating issue", "id_or_prefix", idOrPrefix)

	fullID, err := s.repo.ResolveID(ctx, idOrPrefix)
	if err != nil {
		return nil, err
	}

	if fullID != idOrPrefix {
		slog.Debug("prefix resolved", "input", idOrPrefix, "resolved", fullID)
	}

	issue, err := s.repo.GetIssue(ctx, fullID)
	if err != nil {
		return nil, err
	}

	var changed []string

	if update.Title != nil {
		issue.Title = *update.Title
		changed = append(changed, "title")
	}

	if update.Description != nil {
		issue.Description = *update.Description
		changed = append(changed, "description")
	}

	if update.Status != nil {
		issue.Status = *update.Status
		changed = append(changed, "status")
	}

	if update.Type != nil {
		issue.Type = *update.Type
		changed = append(changed, "type")
	}

	if update.Priority != nil {
		issue.Priority = update.Priority
		changed = append(changed, "priority")
	}

	if update.Assignee != nil {
		issue.Assignee = *update.Assignee
		changed = append(changed, "assignee")
	}

	if update.EstimateMins != nil {
		issue.EstimateMins = *update.EstimateMins
		changed = append(changed, "estimate_mins")
	}

	if update.ParentID != nil {
		issue.ParentID = update.ParentID
		changed = append(changed, "parent_id")
	}

	if update.DeferUntil != nil {
		issue.DeferUntil = update.DeferUntil
		changed = append(changed, "defer_until")
	}

	if update.DueAt != nil {
		issue.DueAt = update.DueAt
		changed = append(changed, "due_at")
	}

	issue.UpdatedAt = time.Now()

	if issue.Status == domain.StatusClosed && issue.ClosedAt == nil {
		now := time.Now()
		issue.ClosedAt = &now
		changed = append(changed, "closed_at")
	}

	if update.Labels != nil {
		issue.Labels = *update.Labels
		changed = append(changed, "labels")
	}

	slog.Debug("updating issue", "id", fullID, "fields_changed", changed)

	if err := s.repo.UpdateIssue(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	slog.Debug("issue updated", "id", fullID, "fields_changed", changed)

	return issue, nil
}

func (s *Service) CloseIssue(ctx context.Context, idOrPrefix string) (*domain.Issue, bool, error) {
	slog.Debug("closing issue", "id_or_prefix", idOrPrefix)

	fullID, err := s.repo.ResolveID(ctx, idOrPrefix)
	if err != nil {
		return nil, false, err
	}

	if fullID != idOrPrefix {
		slog.Debug("prefix resolved", "input", idOrPrefix, "resolved", fullID)
	}

	issue, err := s.repo.GetIssue(ctx, fullID)
	if err != nil {
		return nil, false, err
	}

	changed := issue.Status != domain.StatusClosed
	if changed {
		now := time.Now()
		issue.Status = domain.StatusClosed
		issue.ClosedAt = &now
		issue.UpdatedAt = now

		if err := s.repo.CloseIssue(ctx, fullID, now); err != nil {
			return nil, false, fmt.Errorf("failed to close issue: %w", err)
		}

		slog.Debug("issue closed", "id", fullID)
	} else {
		slog.Debug("issue already closed", "id", fullID)
	}

	issue.Labels, err = s.repo.GetLabels(ctx, fullID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get labels: %w", err)
	}

	issue.Dependencies, err = s.repo.GetDependencies(ctx, fullID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get dependencies: %w", err)
	}

	issue.Comments, err = s.repo.GetComments(ctx, fullID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get comments: %w", err)
	}

	return issue, changed, nil
}

func (s *Service) ReopenIssue(ctx context.Context, idOrPrefix string) (*domain.Issue, bool, error) {
	slog.Debug("reopening issue", "id_or_prefix", idOrPrefix)

	fullID, err := s.repo.ResolveID(ctx, idOrPrefix)
	if err != nil {
		return nil, false, err
	}

	if fullID != idOrPrefix {
		slog.Debug("prefix resolved", "input", idOrPrefix, "resolved", fullID)
	}

	issue, err := s.repo.GetIssue(ctx, fullID)
	if err != nil {
		return nil, false, err
	}

	changed := issue.Status == domain.StatusClosed
	if changed {
		now := time.Now()
		issue.Status = domain.StatusOpen
		issue.ClosedAt = nil
		issue.UpdatedAt = now

		if err := s.repo.ReopenIssue(ctx, fullID, now); err != nil {
			return nil, false, fmt.Errorf("failed to reopen issue: %w", err)
		}

		slog.Debug("issue reopened", "id", fullID)
	} else {
		slog.Debug("issue already open", "id", fullID)
	}

	issue.Labels, err = s.repo.GetLabels(ctx, fullID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get labels: %w", err)
	}

	issue.Dependencies, err = s.repo.GetDependencies(ctx, fullID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get dependencies: %w", err)
	}

	issue.Comments, err = s.repo.GetComments(ctx, fullID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get comments: %w", err)
	}

	return issue, changed, nil
}

func (s *Service) ReadyIssues(ctx context.Context, sort string, limit int) ([]domain.Issue, error) {
	if sort == "" {
		sort = "priority"
	}
	if limit <= 0 {
		limit = 20
	}

	slog.Debug("listing ready issues", "sort", sort, "limit", limit)

	issues, err := s.repo.ReadyIssues(ctx, sort, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list ready issues: %w", err)
	}

	if err := s.populateRelations(ctx, issues); err != nil {
		return nil, err
	}

	slog.Debug("ready issues listed", "count", len(issues))

	return issues, nil
}

func (s *Service) UpcomingIssues(ctx context.Context, days int, assignee string) ([]domain.Issue, error) {
	if days <= 0 {
		days = 15
	}

	slog.Debug("listing upcoming issues", "days", days, "assignee", assignee)

	issues, err := s.repo.UpcomingIssues(ctx, time.Now(), days, assignee)
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming issues: %w", err)
	}

	if err := s.populateRelations(ctx, issues); err != nil {
		return nil, err
	}

	slog.Debug("upcoming issues listed", "count", len(issues))

	return issues, nil
}

func (s *Service) AddDependency(ctx context.Context, blockerIDOrPrefix, blockedIDOrPrefix string) (string, string, error) {
	slog.Debug("adding dependency", "blocker", blockerIDOrPrefix, "blocked", blockedIDOrPrefix)

	blockerID, err := s.repo.ResolveID(ctx, blockerIDOrPrefix)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve blocker: %w", err)
	}

	blockedID, err := s.repo.ResolveID(ctx, blockedIDOrPrefix)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve blocked: %w", err)
	}

	if blockerID == blockedID {
		return "", "", fmt.Errorf("an issue may not block itself")
	}

	dep := domain.Dependency{
		IssueID:     blockedID,
		DependsOnID: blockerID,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.AddDependency(ctx, dep); err != nil {
		return "", "", fmt.Errorf("failed to add dependency: %w", err)
	}

	hasCycle, cycle, err := s.detectCycle(ctx, blockedID)
	if err != nil {
		if _, err := s.repo.RemoveDependency(ctx, blockedID, blockerID); err != nil {
			return "", "", fmt.Errorf("failed to check for cycles and remove dependency: %w", err)
		}
		return "", "", fmt.Errorf("failed to check for cycles: %w", err)
	}

	if hasCycle {
		if _, err := s.repo.RemoveDependency(ctx, blockedID, blockerID); err != nil {
			return "", "", fmt.Errorf("cycle detected and failed to remove dependency: %w", err)
		}
		return "", "", fmt.Errorf("dependency would create a cycle: %s", strings.Join(cycle, " -> "))
	}

	slog.Debug("dependency added", "blocker", blockerID, "blocked", blockedID)

	return blockerID, blockedID, nil
}

func (s *Service) RemoveDependency(ctx context.Context, blockerIDOrPrefix, blockedIDOrPrefix string) (string, string, bool, error) {
	slog.Debug("removing dependency", "blocker", blockerIDOrPrefix, "blocked", blockedIDOrPrefix)

	blockerID, err := s.repo.ResolveID(ctx, blockerIDOrPrefix)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to resolve blocker: %w", err)
	}

	blockedID, err := s.repo.ResolveID(ctx, blockedIDOrPrefix)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to resolve blocked: %w", err)
	}

	changed, err := s.repo.RemoveDependency(ctx, blockedID, blockerID)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to remove dependency: %w", err)
	}

	slog.Debug("dependency removed", "blocker", blockerID, "blocked", blockedID, "changed", changed)

	return blockerID, blockedID, changed, nil
}

func (s *Service) AddComment(ctx context.Context, idOrPrefix string, author string, body string) (*domain.Comment, error) {
	slog.Debug("adding comment", "id_or_prefix", idOrPrefix)

	fullID, err := s.repo.ResolveID(ctx, idOrPrefix)
	if err != nil {
		return nil, err
	}

	if fullID != idOrPrefix {
		slog.Debug("prefix resolved", "input", idOrPrefix, "resolved", fullID)
	}

	comment := &domain.Comment{
		IssueID:   fullID,
		Author:    author,
		Body:      body,
		CreatedAt: time.Now(),
	}

	if err := s.repo.AddComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	slog.Debug("comment added", "id", comment.ID, "issue_id", fullID)

	return comment, nil
}

func (s *Service) detectCycle(ctx context.Context, startID string) (bool, []string, error) {
	deps, err := s.repo.GetDependencyGraph(ctx)
	if err != nil {
		return false, nil, err
	}

	graph := map[string][]string{}
	for _, d := range deps {
		graph[d.IssueID] = append(graph[d.IssueID], d.DependsOnID)
	}

	hasCycle, cycle := dfs.DetectCycle(graph, startID)

	return hasCycle, cycle, nil
}

func (s *Service) populateRelations(ctx context.Context, issues []domain.Issue) error {
	for i := range issues {
		id := issues[i].ID

		labels, err := s.repo.GetLabels(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get labels for %s: %w", id, err)
		}
		issues[i].Labels = labels

		deps, err := s.repo.GetDependencies(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get dependencies for %s: %w", id, err)
		}
		issues[i].Dependencies = deps

		comments, err := s.repo.GetComments(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get comments for %s: %w", id, err)
		}
		issues[i].Comments = comments
	}

	return nil
}

func NewService(repo repo.Repo, imp importer.Importer, prefix string) *Service {
	return &Service{
		repo:   repo,
		imp:    imp,
		prefix: prefix,
	}
}
