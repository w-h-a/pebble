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

func setupService(t *testing.T) *Service {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "pebble.db")

	r, err := sqlite.NewRepo(repo.WithLocation(dbPath))
	require.NoError(t, err)

	t.Cleanup(func() { r.Close() })

	return NewService(r, "test")
}
