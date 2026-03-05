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
