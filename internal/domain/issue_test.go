package domain

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetDefaults_Empty(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	before := time.Now()

	i := &Issue{}

	// Act
	i.SetDefaults()

	after := time.Now()

	// Assert
	require.Equal(t, StatusOpen, i.Status)
	require.Equal(t, TypeTask, i.Type)
	require.Equal(t, 2, *i.Priority)
	require.False(t, i.CreatedAt.Before(before))
	require.True(t, i.CreatedAt.Before(after))
	require.False(t, i.UpdatedAt.Before(before))
	require.True(t, i.UpdatedAt.Before(after))
}

func TestSetDefaults_PreservesExisting(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	p := 1

	i := &Issue{
		Status:    StatusInProgress,
		Type:      TypeBug,
		Priority:  &p,
		CreatedAt: ts,
		UpdatedAt: ts,
	}

	// Act
	i.SetDefaults()

	// Assert
	require.Equal(t, StatusInProgress, i.Status)
	require.Equal(t, TypeBug, i.Type)
	require.Equal(t, 1, *i.Priority)
	require.Equal(t, ts, i.CreatedAt)
	require.Equal(t, ts, i.UpdatedAt)
}

func TestDescendants_FromEpic(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	e1 := "e1"
	c1 := "c1"
	issues := []Issue{
		{ID: "e1", Title: "Epic"},
		{ID: "c1", Title: "Grouping chore", ParentID: &e1},
		{ID: "t1", Title: "Sub-task 1", ParentID: &c1},
		{ID: "t2", Title: "Sub-task 2", ParentID: &c1},
		{ID: "f1", Title: "Direct child 1", ParentID: &e1},
		{ID: "f2", Title: "Direct child 2", ParentID: &e1},
		{ID: "u1", Title: "Unrelated"},
	}

	// Act
	got := Descendants(issues, "e1")

	// Assert
	require.Len(t, got, 5)
	ids := map[string]bool{}
	for _, iss := range got {
		ids[iss.ID] = true
	}
	require.True(t, ids["c1"])
	require.True(t, ids["t1"])
	require.True(t, ids["t2"])
	require.True(t, ids["f1"])
	require.True(t, ids["f2"])
	require.False(t, ids["e1"])
	require.False(t, ids["u1"])
}

func TestDescendants_FromMiddleNode(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	e1 := "e1"
	c1 := "c1"
	issues := []Issue{
		{ID: "e1", Title: "Epic"},
		{ID: "c1", Title: "Grouping chore", ParentID: &e1},
		{ID: "t1", Title: "Sub-task 1", ParentID: &c1},
		{ID: "t2", Title: "Sub-task 2", ParentID: &c1},
		{ID: "f1", Title: "Direct child 1", ParentID: &e1},
		{ID: "f2", Title: "Direct child 2", ParentID: &e1},
		{ID: "u1", Title: "Unrelated"},
	}

	// Act
	got := Descendants(issues, "c1")

	// Assert
	require.Len(t, got, 2)
	ids := map[string]bool{}
	for _, iss := range got {
		ids[iss.ID] = true
	}
	require.True(t, ids["t1"])
	require.True(t, ids["t2"])
}

func TestDescendants_UnknownRoot(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	e1 := "e1"
	issues := []Issue{
		{ID: "e1", Title: "Epic"},
		{ID: "c1", Title: "Child", ParentID: &e1},
	}

	// Act
	got := Descendants(issues, "unknown")

	// Assert
	require.Empty(t, got)
}

func TestDescendants_FlatList(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	issues := []Issue{
		{ID: "a", Title: "Alpha"},
		{ID: "b", Title: "Beta"},
		{ID: "c", Title: "Charlie"},
	}

	// Act
	got := Descendants(issues, "a")

	// Assert
	require.Empty(t, got)
}

func TestListFilterValidate_SinceWithClosed(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := ListFilter{Since: &since, Status: string(StatusClosed)}

	// Act
	err := f.Validate()

	// Assert
	require.NoError(t, err)
}

func TestListFilterValidate_SinceWithOpen(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := ListFilter{Since: &since, Status: string(StatusOpen)}

	// Act
	err := f.Validate()

	// Assert
	require.EqualError(t, err, "filtering by 'since' is only supported when status is 'closed'")
}

func TestListFilterValidate_SinceWithEmptyStatus(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := ListFilter{Since: &since}

	// Act
	err := f.Validate()

	// Assert
	require.EqualError(t, err, "filtering by 'since' is only supported when status is 'closed'")
}

func TestListFilterValidate_NilSince(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	f := ListFilter{}

	// Act
	err := f.Validate()

	// Assert
	require.NoError(t, err)
}

func TestDeleteFilterValidate_ClosedBeforeSet(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	f := DeleteFilter{ClosedBefore: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}

	// Act
	err := f.Validate()

	// Assert
	require.NoError(t, err)
}

func TestDeleteFilterValidate_ClosedBeforeZero(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	f := DeleteFilter{}

	// Act
	err := f.Validate()

	// Assert
	require.EqualError(t, err, "deleting requires a 'closed-before' filter")
}
