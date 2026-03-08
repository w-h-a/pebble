package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/w-h-a/bees/internal/client/importer/noop"
	"github.com/w-h-a/bees/internal/client/repo"
	"github.com/w-h-a/bees/internal/client/repo/sqlite"
	"github.com/w-h-a/bees/internal/domain"
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
	require.Contains(t, id, "test-")
	require.Len(t, id, 9) // "test-" (5) + 4-char hash

	// Act: get by id
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)

	// Assert: full issue assembled
	require.Equal(t, id, got.ID)
	require.Equal(t, "My first task", got.Title)
	require.Equal(t, "Some details", got.Description)
	require.Equal(t, domain.StatusOpen, got.Status)
	require.Equal(t, domain.TypeTask, got.Type)
	require.Equal(t, 2, *got.Priority)
	require.Equal(t, []string{"backend", "v1"}, got.Labels)
	require.False(t, got.CreatedAt.IsZero())
	require.False(t, got.UpdatedAt.IsZero())
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
	require.Equal(t, id, got.ID)
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
	require.Error(t, err)
	require.Contains(t, err.Error(), "no issue found")
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
	require.NoError(t, err)
	require.Equal(t, domain.StatusOpen, got.Status)
	require.Equal(t, domain.TypeTask, got.Type)
	require.Equal(t, 2, *got.Priority)
}

func TestCreateIssue_RejectsPriorityBelowRange(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()
	p := -1

	// Act
	_, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Bad priority", Priority: &p})

	// Assert
	require.Error(t, err)
}

func TestCreateIssue_RejectsPriorityAboveRange(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()
	p := 5

	// Act
	_, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Bad priority", Priority: &p})

	// Assert
	require.Error(t, err)
}

func TestCreateIssue_AcceptsBoundaryPriorities(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act & Assert: priority 0 succeeds
	p0 := 0
	_, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Priority zero", Priority: &p0})
	require.NoError(t, err)

	// Act & Assert: priority 4 succeeds
	p4 := 4
	_, err = svc.CreateIssue(ctx, &domain.Issue{Title: "Priority four", Priority: &p4})
	require.NoError(t, err)
}

func TestCreateIssue_RejectsEmptyTitle(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act
	_, err := svc.CreateIssue(ctx, &domain.Issue{Title: ""})

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "title may not be empty")
}

func TestCreateIssue_RejectsWhitespaceOnlyTitle(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act
	_, err := svc.CreateIssue(ctx, &domain.Issue{Title: "   "})

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "title may not be empty")
}

func TestCreateIssue_AcceptsSingleCharTitle(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act
	id, err := svc.CreateIssue(ctx, &domain.Issue{Title: "x"})

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, id)
}

func TestCreateIssue_EmptyLabelIgnored(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act: create with a single empty-string label
	issue := &domain.Issue{Title: "Empty label create", Labels: []string{""}}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Assert: no labels persisted
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Empty(t, got.Labels)
}

func TestCreateIssue_FiltersEmptyLabels(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act: create with mix of valid and empty labels
	issue := &domain.Issue{Title: "Mixed label create", Labels: []string{"valid", "", " "}}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Assert: only "valid" persisted
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Equal(t, []string{"valid"}, got.Labels)
}

func TestCreateIssue_RejectsNonexistentParent(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act
	badParent := "nonexistent-id"
	_, err := svc.CreateIssue(ctx, &domain.Issue{
		Title:    "Child with bad parent",
		ParentID: &badParent,
	})

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "nonexistent-id")
	require.Contains(t, err.Error(), "not found")
}

func TestCreateIssue_AcceptsValidParent(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	parentID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Parent"})
	require.NoError(t, err)

	// Act
	childID, err := svc.CreateIssue(ctx, &domain.Issue{
		Title:    "Child",
		ParentID: &parentID,
	})
	require.NoError(t, err)

	// Assert
	got, err := svc.GetIssue(ctx, childID)
	require.NoError(t, err)
	require.NotNil(t, got.ParentID)
	require.Equal(t, parentID, *got.ParentID)
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
	require.Len(t, issues, 1)
	require.Equal(t, "Open task", issues[0].Title)
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
	require.Len(t, issues, 1)
	require.Equal(t, "Tagged", issues[0].Title)
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
	require.Len(t, issues, 2)
}

func TestListIssues_PopulatesRelations(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Hydration test", Labels: []string{"backend", "v1"}}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	blocker := &domain.Issue{Title: "Blocker"}
	blockerID, err := svc.CreateIssue(ctx, blocker)
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, blockerID, id)
	require.NoError(t, err)

	_, err = svc.AddComment(ctx, id, "wes", "test comment")
	require.NoError(t, err)

	// Act
	issues, err := svc.ListIssues(ctx, domain.ListFilter{})
	require.NoError(t, err)

	// Assert: find the hydration test issue
	var got *domain.Issue
	for i := range issues {
		if issues[i].ID == id {
			got = &issues[i]
			break
		}
	}
	require.NotNil(t, got, "expected issue not found in list")

	// Assert: all three relations populated
	require.Equal(t, []string{"backend", "v1"}, got.Labels)
	require.Len(t, got.Dependencies, 1)
	require.Equal(t, blockerID, got.Dependencies[0].DependsOnID)
	require.Len(t, got.Comments, 1)
	require.Equal(t, "test comment", got.Comments[0].Body)
}

func TestSearchIssues_ByTitle(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	_, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Fix login bug"})
	require.NoError(t, err)

	_, err = svc.CreateIssue(ctx, &domain.Issue{Title: "Add signup page"})
	require.NoError(t, err)

	// Act
	issues, err := svc.SearchIssues(ctx, "login", 0)
	require.NoError(t, err)

	// Assert
	require.Len(t, issues, 1)
	require.Equal(t, "Fix login bug", issues[0].Title)
}

func TestSearchIssues_ByDescription(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	_, err := svc.CreateIssue(ctx, &domain.Issue{
		Title:       "Task A",
		Description: "Refactor the authentication module",
	})
	require.NoError(t, err)

	_, err = svc.CreateIssue(ctx, &domain.Issue{Title: "Task B"})
	require.NoError(t, err)

	// Act
	issues, err := svc.SearchIssues(ctx, "authentication", 0)
	require.NoError(t, err)

	// Assert
	require.Len(t, issues, 1)
	require.Equal(t, "Task A", issues[0].Title)
}

func TestSearchIssues_NoResults(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	_, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Something"})
	require.NoError(t, err)

	// Act
	issues, err := svc.SearchIssues(ctx, "nonexistent", 0)
	require.NoError(t, err)

	// Assert
	require.Empty(t, issues)
}

func TestSearchIssues_MatchesBothTitleAndDescription(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	_, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Cache warming"})
	require.NoError(t, err)

	_, err = svc.CreateIssue(ctx, &domain.Issue{
		Title:       "Perf improvement",
		Description: "Add cache layer to reduce latency",
	})
	require.NoError(t, err)

	// Act
	issues, err := svc.SearchIssues(ctx, "cache", 0)
	require.NoError(t, err)

	// Assert: both match
	require.Len(t, issues, 2)
}

func TestSearchIssues_PopulatesRelations(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Searchable hydration", Labels: []string{"api"}}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	blocker := &domain.Issue{Title: "Blocker"}
	blockerID, err := svc.CreateIssue(ctx, blocker)
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, blockerID, id)
	require.NoError(t, err)

	_, err = svc.AddComment(ctx, id, "wes", "search comment")
	require.NoError(t, err)

	// Act
	issues, err := svc.SearchIssues(ctx, "Searchable", 0)
	require.NoError(t, err)

	// Assert: relations populated
	require.Len(t, issues, 1)
	require.Equal(t, []string{"api"}, issues[0].Labels)
	require.Len(t, issues[0].Dependencies, 1)
	require.Equal(t, blockerID, issues[0].Dependencies[0].DependsOnID)
	require.Len(t, issues[0].Comments, 1)
	require.Equal(t, "search comment", issues[0].Comments[0].Body)
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
	require.Equal(t, "Updated title", updated.Title)
	require.Equal(t, 0, *updated.Priority)
	require.Equal(t, []string{"urgent", "backend"}, updated.Labels)

	// Assert: persisted correctly
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Updated title", got.Title)
	require.Equal(t, 0, *got.Priority)
	require.Equal(t, []string{"backend", "urgent"}, got.Labels)
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
	require.Equal(t, domain.StatusClosed, updated.Status)
	require.NotNil(t, updated.ClosedAt)
}

func TestUpdateIssue_ReopenClearsClosedAt(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Will reopen via update"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	closedStatus := domain.StatusClosed
	closed, err := svc.UpdateIssue(ctx, id, domain.IssueUpdate{Status: &closedStatus})
	require.NoError(t, err)
	require.NotNil(t, closed.ClosedAt)

	// Act
	openStatus := domain.StatusOpen
	updated, err := svc.UpdateIssue(ctx, id, domain.IssueUpdate{Status: &openStatus})
	require.NoError(t, err)

	// Assert: returned issue has no ClosedAt
	require.Equal(t, domain.StatusOpen, updated.Status)
	require.Nil(t, updated.ClosedAt)

	// Assert: persisted correctly
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Equal(t, domain.StatusOpen, got.Status)
	require.Nil(t, got.ClosedAt)
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
	require.Equal(t, id, updated.ID)
	require.Equal(t, "Updated via prefix", updated.Title)
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
	require.Equal(t, "New title", updated.Title)
	require.Equal(t, "wes", updated.Assignee)
	require.Equal(t, 1, *updated.Priority)
}

func TestUpdateIssue_RejectsPriorityOutOfRange(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	id, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Valid issue"})
	require.NoError(t, err)

	// Act & Assert: priority -1 fails
	pNeg := -1
	_, err = svc.UpdateIssue(ctx, id, domain.IssueUpdate{Priority: &pNeg})
	require.Error(t, err)

	// Act & Assert: priority 5 fails
	p5 := 5
	_, err = svc.UpdateIssue(ctx, id, domain.IssueUpdate{Priority: &p5})
	require.Error(t, err)

	// Assert: issue unchanged
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Equal(t, 2, *got.Priority)
}

func TestUpdateIssue_EmptyLabelClearsAll(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Label test", Labels: []string{"a", "b"}}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act: update with a single empty string (simulates --label '')
	empty := []string{""}
	_, err = svc.UpdateIssue(ctx, id, domain.IssueUpdate{Labels: &empty})
	require.NoError(t, err)

	// Assert: issue has zero labels, not one empty-string label
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Empty(t, got.Labels)
}

func TestUpdateIssue_FiltersEmptyLabels(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Mixed labels"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act: update with mix of valid and empty labels
	mixed := []string{"x", "", " "}
	_, err = svc.UpdateIssue(ctx, id, domain.IssueUpdate{Labels: &mixed})
	require.NoError(t, err)

	// Assert: only "x" survives
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Equal(t, []string{"x"}, got.Labels)
}

func TestUpdateIssue_RejectsNonexistentParent(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	id, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Existing issue"})
	require.NoError(t, err)

	// Act
	badParent := "nonexistent-id"
	_, err = svc.UpdateIssue(ctx, id, domain.IssueUpdate{ParentID: &badParent})

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "nonexistent-id")
	require.Contains(t, err.Error(), "not found")

	// Assert: parent_id unchanged
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Nil(t, got.ParentID)
}

func TestUpdateIssue_AcceptsValidParent(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	parentID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Parent"})
	require.NoError(t, err)

	childID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Child"})
	require.NoError(t, err)

	// Act
	updated, err := svc.UpdateIssue(ctx, childID, domain.IssueUpdate{ParentID: &parentID})
	require.NoError(t, err)

	// Assert
	require.NotNil(t, updated.ParentID)
	require.Equal(t, parentID, *updated.ParentID)

	// Assert: persisted
	got, err := svc.GetIssue(ctx, childID)
	require.NoError(t, err)
	require.NotNil(t, got.ParentID)
	require.Equal(t, parentID, *got.ParentID)
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
	require.True(t, changed)
	require.Equal(t, domain.StatusClosed, closed.Status)
	require.NotNil(t, closed.ClosedAt)
	require.Equal(t, []string{"bug"}, closed.Labels)

	// Assert: persisted
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Equal(t, domain.StatusClosed, got.Status)
	require.NotNil(t, got.ClosedAt)
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
	require.False(t, changed)
	require.Equal(t, domain.StatusClosed, got.Status)
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
	require.True(t, changed)
	require.Equal(t, domain.StatusOpen, reopened.Status)
	require.Nil(t, reopened.ClosedAt)

	// Assert: persisted
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Equal(t, domain.StatusOpen, got.Status)
	require.Nil(t, got.ClosedAt)
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
	require.False(t, changed)
	require.Equal(t, domain.StatusOpen, got.Status)
}

func TestReadyIssues_ExcludesDeferredAndClosed(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	ready := &domain.Issue{Title: "Ready task"}
	_, err := svc.CreateIssue(ctx, ready)
	require.NoError(t, err)

	deferred := &domain.Issue{Title: "Deferred task"}
	deferredID, err := svc.CreateIssue(ctx, deferred)
	require.NoError(t, err)

	future := time.Now().AddDate(0, 0, 7)
	_, err = svc.UpdateIssue(ctx, deferredID, domain.IssueUpdate{DeferUntil: &future})
	require.NoError(t, err)

	closed := &domain.Issue{Title: "Closed task"}
	closedID, err := svc.CreateIssue(ctx, closed)
	require.NoError(t, err)

	_, _, err = svc.CloseIssue(ctx, closedID)
	require.NoError(t, err)

	// Act
	issues, err := svc.ReadyIssues(ctx, "", 0)
	require.NoError(t, err)

	// Assert: only the undeferred open issue
	require.Len(t, issues, 1)
	require.Equal(t, "Ready task", issues[0].Title)
}

func TestReadyIssues_IncludesPastDeferred(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Was deferred"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	yesterday := time.Now().AddDate(0, 0, -1)
	_, err = svc.UpdateIssue(ctx, id, domain.IssueUpdate{DeferUntil: &yesterday})
	require.NoError(t, err)

	// Act
	issues, err := svc.ReadyIssues(ctx, "", 0)
	require.NoError(t, err)

	// Assert: past-deferred issue is ready
	require.Len(t, issues, 1)
	require.Equal(t, "Was deferred", issues[0].Title)
}

func TestUpcomingIssues_WindowFilter(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	today := time.Now()
	threeDays := today.AddDate(0, 0, 3)
	tenDays := today.AddDate(0, 0, 10)

	todayIssue := &domain.Issue{Title: "Today"}
	todayID, err := svc.CreateIssue(ctx, todayIssue)
	require.NoError(t, err)
	_, err = svc.UpdateIssue(ctx, todayID, domain.IssueUpdate{DeferUntil: &today})
	require.NoError(t, err)

	soonIssue := &domain.Issue{Title: "Soon"}
	soonID, err := svc.CreateIssue(ctx, soonIssue)
	require.NoError(t, err)
	_, err = svc.UpdateIssue(ctx, soonID, domain.IssueUpdate{DeferUntil: &threeDays})
	require.NoError(t, err)

	farIssue := &domain.Issue{Title: "Far out"}
	farID, err := svc.CreateIssue(ctx, farIssue)
	require.NoError(t, err)
	_, err = svc.UpdateIssue(ctx, farID, domain.IssueUpdate{DeferUntil: &tenDays})
	require.NoError(t, err)

	// Act
	issues, err := svc.UpcomingIssues(ctx, 7, "")
	require.NoError(t, err)

	// Assert: today and soon, not far
	require.Len(t, issues, 2)
	require.Equal(t, "Today", issues[0].Title)
	require.Equal(t, "Soon", issues[1].Title)
}

func TestUpcomingIssues_ExcludesNoDefer(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "No defer"}
	_, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act
	issues, err := svc.UpcomingIssues(ctx, 7, "")
	require.NoError(t, err)

	// Assert
	require.Empty(t, issues)
}

func TestUpcomingIssues_FilterByAssignee(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	tomorrow := time.Now().AddDate(0, 0, 1)
	wes := "wes"
	other := "other"

	a := &domain.Issue{Title: "Wes task"}
	aID, err := svc.CreateIssue(ctx, a)
	require.NoError(t, err)

	_, err = svc.UpdateIssue(ctx, aID, domain.IssueUpdate{DeferUntil: &tomorrow, Assignee: &wes})
	require.NoError(t, err)

	b := &domain.Issue{Title: "Other task"}
	bID, err := svc.CreateIssue(ctx, b)
	require.NoError(t, err)

	_, err = svc.UpdateIssue(ctx, bID, domain.IssueUpdate{DeferUntil: &tomorrow, Assignee: &other})
	require.NoError(t, err)

	// Act
	issues, err := svc.UpcomingIssues(ctx, 7, "wes")
	require.NoError(t, err)

	// Assert
	require.Len(t, issues, 1)
	require.Equal(t, "Wes task", issues[0].Title)
}

func TestUpcomingIssues_ExcludesClosed(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	tomorrow := time.Now().AddDate(0, 0, 1)

	issue := &domain.Issue{Title: "Closed before defer"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	_, err = svc.UpdateIssue(ctx, id, domain.IssueUpdate{DeferUntil: &tomorrow})
	require.NoError(t, err)

	_, _, err = svc.CloseIssue(ctx, id)
	require.NoError(t, err)

	// Act
	issues, err := svc.UpcomingIssues(ctx, 7, "")
	require.NoError(t, err)

	// Assert: closed issue excluded despite defer_until in window
	require.Empty(t, issues)
}

func TestUpcomingIssues_IncludesType(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	tomorrow := time.Now().AddDate(0, 0, 1)

	issue := &domain.Issue{Title: "Feature task", Type: domain.TypeFeature}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	_, err = svc.UpdateIssue(ctx, id, domain.IssueUpdate{DeferUntil: &tomorrow})
	require.NoError(t, err)

	// Act
	issues, err := svc.UpcomingIssues(ctx, 7, "")
	require.NoError(t, err)

	// Assert
	require.Len(t, issues, 1)
	require.Equal(t, domain.TypeFeature, issues[0].Type)
}

func TestAddDependency(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	a := &domain.Issue{Title: "Blocker"}
	aID, err := svc.CreateIssue(ctx, a)
	require.NoError(t, err)

	b := &domain.Issue{Title: "Blocked"}
	bID, err := svc.CreateIssue(ctx, b)
	require.NoError(t, err)

	// Act
	blocker, blocked, err := svc.AddDependency(ctx, aID, bID)
	require.NoError(t, err)

	// Assert
	require.Equal(t, aID, blocker)
	require.Equal(t, bID, blocked)

	got, err := svc.GetIssue(ctx, bID)
	require.NoError(t, err)
	require.Len(t, got.Dependencies, 1)
	require.Equal(t, aID, got.Dependencies[0].DependsOnID)
}

func TestAddDependency_SelfBlock(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	a := &domain.Issue{Title: "Self"}
	aID, err := svc.CreateIssue(ctx, a)
	require.NoError(t, err)

	// Act
	_, _, err = svc.AddDependency(ctx, aID, aID)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "may not block itself")
}

func TestAddDependency_CycleDetected(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	a := &domain.Issue{Title: "A"}
	aID, err := svc.CreateIssue(ctx, a)
	require.NoError(t, err)

	b := &domain.Issue{Title: "B"}
	bID, err := svc.CreateIssue(ctx, b)
	require.NoError(t, err)

	c := &domain.Issue{Title: "C"}
	cID, err := svc.CreateIssue(ctx, c)
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, aID, bID) // A blocks B
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, bID, cID) // B blocks C
	require.NoError(t, err)

	// Snapshot graph before cycle attempt
	beforeA, err := svc.GetIssue(ctx, aID)
	require.NoError(t, err)
	beforeB, err := svc.GetIssue(ctx, bID)
	require.NoError(t, err)
	beforeC, err := svc.GetIssue(ctx, cID)
	require.NoError(t, err)

	// Act: C blocks A → cycle
	_, _, err = svc.AddDependency(ctx, cID, aID)

	// Assert: error returned
	require.Error(t, err)
	require.Contains(t, err.Error(), "would create a cycle")

	// Assert: edge was rolled back
	afterA, err := svc.GetIssue(ctx, aID)
	require.NoError(t, err)
	afterB, err := svc.GetIssue(ctx, bID)
	require.NoError(t, err)
	afterC, err := svc.GetIssue(ctx, cID)
	require.NoError(t, err)

	require.Equal(t, len(beforeA.Dependencies), len(afterA.Dependencies))
	require.Equal(t, len(beforeB.Dependencies), len(afterB.Dependencies))
	require.Equal(t, len(beforeC.Dependencies), len(afterC.Dependencies))

	require.Empty(t, afterA.Dependencies)
}

func TestRemoveDependency(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	a := &domain.Issue{Title: "Blocker"}
	aID, err := svc.CreateIssue(ctx, a)
	require.NoError(t, err)

	b := &domain.Issue{Title: "Blocked"}
	bID, err := svc.CreateIssue(ctx, b)
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, aID, bID)
	require.NoError(t, err)

	// Act
	blocker, blocked, changed, err := svc.RemoveDependency(ctx, aID, bID)
	require.NoError(t, err)

	// Assert
	require.True(t, changed)
	require.Equal(t, aID, blocker)
	require.Equal(t, bID, blocked)

	got, err := svc.GetIssue(ctx, bID)
	require.NoError(t, err)
	require.Empty(t, got.Dependencies)
}

func TestRemoveDependency_Idempotent(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	a := &domain.Issue{Title: "A"}
	aID, err := svc.CreateIssue(ctx, a)
	require.NoError(t, err)

	b := &domain.Issue{Title: "B"}
	bID, err := svc.CreateIssue(ctx, b)
	require.NoError(t, err)

	// Act: remove non-existent dependency
	_, _, changed, err := svc.RemoveDependency(ctx, aID, bID)
	require.NoError(t, err)

	// Assert
	require.False(t, changed)
}

func TestBuildGraph_FullGraph(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	aID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Alpha"})
	require.NoError(t, err)
	bID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Beta"})
	require.NoError(t, err)
	cID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Charlie"})
	require.NoError(t, err)
	dID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Delta"})
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, aID, bID)
	require.NoError(t, err)
	_, _, err = svc.AddDependency(ctx, cID, dID)
	require.NoError(t, err)

	// Act
	graph, err := svc.BuildGraph(ctx, nil, "")
	require.NoError(t, err)

	// Assert
	require.Len(t, graph.Nodes, 4)
	require.Len(t, graph.Edges, 2)
	require.Contains(t, graph.Nodes, aID)
	require.Contains(t, graph.Nodes, bID)
	require.Contains(t, graph.Nodes, cID)
	require.Contains(t, graph.Nodes, dID)
}

func TestBuildGraph_ScopedToID(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	aID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Alpha"})
	require.NoError(t, err)
	bID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Beta"})
	require.NoError(t, err)
	cID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Charlie"})
	require.NoError(t, err)
	dID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Delta"})
	require.NoError(t, err)
	eID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Echo"})
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, aID, bID)
	require.NoError(t, err)
	_, _, err = svc.AddDependency(ctx, bID, cID)
	require.NoError(t, err)
	_, _, err = svc.AddDependency(ctx, dID, eID)
	require.NoError(t, err)

	// Act
	graph, err := svc.BuildGraph(ctx, &aID, "")
	require.NoError(t, err)

	// Assert
	require.Len(t, graph.Nodes, 3)
	require.Contains(t, graph.Nodes, aID)
	require.Contains(t, graph.Nodes, bID)
	require.Contains(t, graph.Nodes, cID)
	require.NotContains(t, graph.Nodes, dID)
	require.NotContains(t, graph.Nodes, eID)
	require.Len(t, graph.Edges, 2)
}

func TestBuildGraph_EmptyGraph(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act
	graph, err := svc.BuildGraph(ctx, nil, "")
	require.NoError(t, err)

	// Assert
	require.Empty(t, graph.Nodes)
	require.Empty(t, graph.Edges)
}

func TestBuildGraph_NodeFields(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(24 * time.Hour)

	aID, err := svc.CreateIssue(ctx, &domain.Issue{
		Title:        "Alpha",
		Type:         domain.TypeFeature,
		DeferUntil:   &now,
		EstimateMins: 30,
	})
	require.NoError(t, err)
	bID, err := svc.CreateIssue(ctx, &domain.Issue{
		Title:        "Beta",
		Type:         domain.TypeChore,
		EstimateMins: 20,
	})
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, aID, bID)
	require.NoError(t, err)

	// Act
	graph, err := svc.BuildGraph(ctx, nil, "")
	require.NoError(t, err)

	// Assert
	nodeA := graph.Nodes[aID]
	require.Equal(t, domain.TypeFeature, nodeA.Type)
	require.NotNil(t, nodeA.DeferUntil)
	require.Equal(t, now, *nodeA.DeferUntil)
	require.Equal(t, 30, nodeA.EstimateMins)

	nodeB := graph.Nodes[bID]
	require.Equal(t, domain.TypeChore, nodeB.Type)
	require.Nil(t, nodeB.DeferUntil)
	require.Equal(t, 20, nodeB.EstimateMins)
}

func TestBuildGraph_DefaultExcludesClosed(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	aID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Open issue"})
	require.NoError(t, err)
	bID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Closed issue"})
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, aID, bID)
	require.NoError(t, err)

	_, _, err = svc.CloseIssue(ctx, bID)
	require.NoError(t, err)

	// Act
	graph, err := svc.BuildGraph(ctx, nil, "")
	require.NoError(t, err)

	// Assert: closed node excluded
	require.Len(t, graph.Nodes, 1)
	require.Contains(t, graph.Nodes, aID)
	require.NotContains(t, graph.Nodes, bID)
	require.Empty(t, graph.Edges)
}

func TestBuildGraph_StatusFilter(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	aID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Open issue"})
	require.NoError(t, err)
	bID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Closed issue"})
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, aID, bID)
	require.NoError(t, err)

	_, _, err = svc.CloseIssue(ctx, bID)
	require.NoError(t, err)

	// Act: filter by "closed"
	graph, err := svc.BuildGraph(ctx, nil, "closed")
	require.NoError(t, err)

	// Assert: only closed node
	require.Len(t, graph.Nodes, 1)
	require.Contains(t, graph.Nodes, bID)
	require.NotContains(t, graph.Nodes, aID)
	require.Empty(t, graph.Edges)
}

func TestBuildGraph_StatusAll(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	aID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Open issue"})
	require.NoError(t, err)
	bID, err := svc.CreateIssue(ctx, &domain.Issue{Title: "Closed issue"})
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, aID, bID)
	require.NoError(t, err)

	_, _, err = svc.CloseIssue(ctx, bID)
	require.NoError(t, err)

	// Act: pass "all"
	graph, err := svc.BuildGraph(ctx, nil, "all")
	require.NoError(t, err)

	// Assert: both nodes present
	require.Len(t, graph.Nodes, 2)
	require.Contains(t, graph.Nodes, aID)
	require.Contains(t, graph.Nodes, bID)
	require.Len(t, graph.Edges, 1)
}

func TestReadyIssues_ExcludesBlocked(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	blocker := &domain.Issue{Title: "Blocker"}
	blockerID, err := svc.CreateIssue(ctx, blocker)
	require.NoError(t, err)

	blocked := &domain.Issue{Title: "Blocked"}
	blockedID, err := svc.CreateIssue(ctx, blocked)
	require.NoError(t, err)

	_, _, err = svc.AddDependency(ctx, blockerID, blockedID)
	require.NoError(t, err)

	// Act
	issues, err := svc.ReadyIssues(ctx, "", 0)
	require.NoError(t, err)

	// Assert: only the blocker is ready
	require.Len(t, issues, 1)
	require.Equal(t, "Blocker", issues[0].Title)
}

func TestAddComment(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Commentable"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act
	comment, err := svc.AddComment(ctx, id, "wes", "looks good")
	require.NoError(t, err)

	// Assert: returned comment
	require.Equal(t, id, comment.IssueID)
	require.Equal(t, "wes", comment.Author)
	require.Equal(t, "looks good", comment.Body)
	require.NotZero(t, comment.ID)
	require.False(t, comment.CreatedAt.IsZero())

	// Assert: visible in GetIssue
	got, err := svc.GetIssue(ctx, id)
	require.NoError(t, err)
	require.Len(t, got.Comments, 1)
	require.Equal(t, "looks good", got.Comments[0].Body)
}

func TestAddComment_PrefixResolution(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	issue := &domain.Issue{Title: "Prefix comment test"}
	id, err := svc.CreateIssue(ctx, issue)
	require.NoError(t, err)

	// Act: comment by prefix
	partial := id[:len(id)-1]
	comment, err := svc.AddComment(ctx, partial, "", "via prefix")
	require.NoError(t, err)

	// Assert
	require.Equal(t, id, comment.IssueID)
	require.Equal(t, "via prefix", comment.Body)
}

func TestAddComment_NotFound(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// Arrange
	svc := setupService(t)
	ctx := context.Background()

	// Act
	_, err := svc.AddComment(ctx, "test-zzzz", "", "orphan")

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "no issue found")
}

func setupService(t *testing.T) *Service {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "bees.db")

	r, err := sqlite.NewRepo(repo.WithLocation(dbPath))
	require.NoError(t, err)

	i, err := noop.NewImporter()
	require.NoError(t, err)

	t.Cleanup(func() { r.Close() })

	return NewService(r, i, "test")
}
