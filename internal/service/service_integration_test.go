package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/w-h-a/pebble/internal/client/repo"
	"github.com/w-h-a/pebble/internal/client/repo/sqlite"
	"github.com/w-h-a/pebble/internal/domain"
)

func TestCreateAndGetIssue(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act: create
	issue := &domain.Issue{
		Title:       "My first task",
		Description: "Some details",
		Type:        domain.TypeTask,
		Labels:      []string{"backend", "v1"},
	}

	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Assert: create
	assert.Contains(t, id, "test-")
	assert.Len(t, id, 9) // "test-" (5) + 4-char hash

	// Act: get by id
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)

	// Assert: full issue assembled
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "My first task", got.Title)
	assert.Equal(t, "Some details", got.Description)
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.Equal(t, domain.TypeTask, got.Type)
	assert.Equal(t, 2, *got.Priority)
	assert.Equal(t, []string{"backend", "v1"}, got.Labels)
	assert.False(t, got.CreatedAt.IsZero())
	assert.False(t, got.UpdatedAt.IsZero())
}

func TestGetIssue_PrefixResolution(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Prefix test"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act: resolve by prefix (test-a1b -> test-a1b2)
	partial := id[:len(id)-1]
	got, err := svc.GetIssue(ctx, partial)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, id, got.ID)
}

func TestGetIssue_NotFound(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act
	_, err := svc.GetIssue(ctx, "text-zzzz")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no issue found")
}

func TestCreateIssue_DefaultsApplied(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act: create with minimal fields
	issue := &domain.Issue{Title: "Bare minimum"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Assert: defaults
	got, err := svc.GetIssue(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.Equal(t, domain.TypeTask, got.Type)
	assert.Equal(t, 2, *got.Priority)
}

func TestListIssues_DefaultsToOpen(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	open := &domain.Issue{Title: "Open task"}
	_, err := svc.CreateIssue(ctx, open)
	require.NoError(t, err)

	closed := &domain.Issue{Title: "Closed task"}
	closedID, err := svc.CreateIssue(ctx, closed)
	require.NoError(t, err)

	closedStatus := domain.StatusClosed
	_, err = svc.UpdateIssue(ctx, closedID, domain.IssueUpdate{Status: &closedStatus})
	require.NoError(t, err)

	// Act
	issues, err := svc.ListIssues(ctx, domain.ListFilter{})
	require.NoError(t, err)

	// Assert: only the open issue
	assert.Len(t, issues, 1)
	assert.Equal(t, "Open task", issues[0].Title)
}

func TestListIssues_FilterByLabel(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	tagged := &domain.Issue{Title: "Tagged", Labels: []string{"urgent"}}
	_, err := svc.CreateIssue(ctx, tagged)
	require.NoError(t, err)

	untagged := &domain.Issue{Title: "Untagged"}
	_, err = svc.CreateIssue(ctx, untagged)
	require.NoError(t, err)

	// Act
	issues, err := svc.ListIssues(ctx, domain.ListFilter{Label: "urgent"})
	require.NoError(t, err)

	// Assert
	assert.Len(t, issues, 1)
	assert.Equal(t, "Tagged", issues[0].Title)
}

func TestListIssues_StatusAll(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	open := &domain.Issue{Title: "Open"}
	_, err := svc.CreateIssue(ctx, open)
	require.NoError(t, err)

	closed := &domain.Issue{Title: "Closed"}
	closedID, err := svc.CreateIssue(ctx, closed)
	require.NoError(t, err)

	closedStatus := domain.StatusClosed
	_, err = svc.UpdateIssue(ctx, closedID, domain.IssueUpdate{Status: &closedStatus})
	require.NoError(t, err)

	// Act
	issues, err := svc.ListIssues(ctx, domain.ListFilter{Status: "all"})
	require.NoError(t, err)

	// Assert: both
	assert.Len(t, issues, 2)
}

func TestUpdateIssue_AppliesChanges(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	svc := setupService(t)
	ctx := context.Background()

	// Arrange
	issue := &domain.Issue{Title: "Original title", Labels: []string{"old"}}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act: update title, priority, and labels
	newTitle := "Updated title"
	newPriority := 0
	newLabels := []string{"urgent", "backend"}
	updated, err := svc.UpdateIssue(ctx, id, domain.IssueUpdate{
		Title:    &newTitle,
		Priority: &newPriority,
		Labels:   &newLabels,
	})
	require.NoError(t, err)

	// Assert: returned issue reflects changes
	assert.Equal(t, "Updated title", updated.Title)
	assert.Equal(t, 0, *updated.Priority)
	assert.Equal(t, []string{"urgent", "backend"}, updated.Labels)

	// Assert: persisted correctly
	got, err := svc.GetIssue(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "Updated title", got.Title)
	assert.Equal(t, 0, *got.Priority)
	assert.Equal(t, []string{"backend", "urgent"}, got.Labels)
}

func TestUpdateIssue_ClosedSetsClosedAt(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Will close"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act
	closedStatus := domain.StatusClosed
	updated, err := svc.UpdateIssue(ctx, id, domain.IssueUpdate{Status: &closedStatus})
	require.NoError(t, err)

	// Assert
	assert.Equal(t, domain.StatusClosed, updated.Status)
	assert.NotNil(t, updated.ClosedAt)
}

func TestUpdateIssue_PrefixResolution(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Prefix update test"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act: update by prefix
	partial := id[:len(id)-1]
	newTitle := "Updated via prefix"
	updated, err := svc.UpdateIssue(ctx, partial, domain.IssueUpdate{Title: &newTitle})
	require.NoError(t, err)

	// Assert
	assert.Equal(t, id, updated.ID)
	assert.Equal(t, "Updated via prefix", updated.Title)
}

func TestUpdateIssue_UnchangedFieldsPreserved(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	p := 1
	issue := &domain.Issue{
		Title:    "Original",
		Assignee: "wes",
		Priority: &p,
	}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act: update only title
	newTitle := "New title"
	updated, err := svc.UpdateIssue(ctx, id, domain.IssueUpdate{Title: &newTitle})
	require.NoError(t, err)

	// Assert: other fields untouched
	assert.Equal(t, "New title", updated.Title)
	assert.Equal(t, "wes", updated.Assignee)
	assert.Equal(t, 1, *updated.Priority)
}

func TestCloseIssue(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "To close", Labels: []string{"bug"}}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act
	closed, changed, err := svc.CloseIssue(ctx, id)
	require.NoError(t, err)

	// Assert
	assert.True(t, changed)
	assert.Equal(t, domain.StatusClosed, closed.Status)
	assert.NotNil(t, closed.ClosedAt)
	assert.Equal(t, []string{"bug"}, closed.Labels)

	// Assert: persisted
	got, err := svc.GetIssue(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, got.Status)
	assert.NotNil(t, got.ClosedAt)
}

func TestCloseIssue_AlreadyClosed(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Already closed"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	_, _, err = svc.CloseIssue(ctx, id)
	require.NoError(t, err)

	// Act: close again
	got, changed, err := svc.CloseIssue(ctx, id)
	require.NoError(t, err)

	// Assert: idempotent
	assert.False(t, changed)
	assert.Equal(t, domain.StatusClosed, got.Status)
}

func TestReopenIssue(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "To reopen"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	_, _, err = svc.CloseIssue(ctx, id)
	require.NoError(t, err)

	// Act
	reopened, changed, err := svc.ReopenIssue(ctx, id)
	require.NoError(t, err)

	// Assert
	assert.True(t, changed)
	assert.Equal(t, domain.StatusOpen, reopened.Status)
	assert.Nil(t, reopened.ClosedAt)

	// Assert: persisted
	got, err := svc.GetIssue(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.Nil(t, got.ClosedAt)
}

func TestReopenIssue_AlreadyOpen(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Already open"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act: reopen without closing first
	got, changed, err := svc.ReopenIssue(ctx, id)
	require.NoError(t, err)

	// Assert: idempotent
	assert.False(t, changed)
	assert.Equal(t, domain.StatusOpen, got.Status)
}

func setupService(t *testing.T) *Service {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "pebble.db")

	r, err := sqlite.NewRepo(repo.WithLocation(dbPath))
	require.NoError(t, err)

	t.Cleanup(func() { r.Close() })

	return NewService(r, "test")
}
