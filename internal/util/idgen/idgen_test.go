package idgen

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerate_Format(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	frozen := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return frozen }
	t.Cleanup(func() { timeNow = time.Now })

	// Act
	id := Generate("pb", "test title", "test desc", 0)

	// Assert
	assert.Contains(t, id, "pb-")
	assert.Len(t, id, 7)
}

func TestGenerate_Deterministic(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	frozen := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return frozen }
	t.Cleanup(func() { timeNow = time.Now })

	// Act
	a := Generate("pb", "title", "desc", 0)
	b := Generate("pb", "title", "desc", 0)

	// Assert
	assert.Equal(t, a, b)
}

func TestGenerate_NonceChangesSuffix(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	frozen := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return frozen }
	t.Cleanup(func() { timeNow = time.Now })

	// Act
	a := Generate("pb", "title", "desc", 0)
	b := Generate("pb", "title", "desc", 1)

	// Assert
	assert.NotEqual(t, a, b)
}

func TestGenerate_DifferentPrefixes(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	frozen := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return frozen }
	t.Cleanup(func() { timeNow = time.Now })

	// Act
	a := Generate("pb", "title", "desc", 0)
	b := Generate("xx", "title", "desc", 0)

	// Assert
	assert.NotEqual(t, a, b)
	assert.True(t, strings.HasPrefix(b, "xx-"))
}
