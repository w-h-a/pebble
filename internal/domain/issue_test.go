package domain

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, StatusOpen, i.Status)
	assert.Equal(t, TypeTask, i.Type)
	assert.Equal(t, 2, *i.Priority)
	assert.False(t, i.CreatedAt.Before(before))
	assert.True(t, i.CreatedAt.Before(after))
	assert.False(t, i.UpdatedAt.Before(before))
	assert.True(t, i.UpdatedAt.Before(after))
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
	assert.Equal(t, StatusInProgress, i.Status)
	assert.Equal(t, TypeBug, i.Type)
	assert.Equal(t, 1, *i.Priority)
	assert.Equal(t, ts, i.CreatedAt)
	assert.Equal(t, ts, i.UpdatedAt)
}
