package dfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectCycle_EmptyGraph(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Act
	hasCycle, cycle := DetectCycle(nil, "a")

	// Assert
	assert.False(t, hasCycle)
	assert.Empty(t, cycle)
}

func TestDetectCycle_LinearChain(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
	}

	// Act
	hasCycle, cycle := DetectCycle(graph, "a")

	// Assert
	assert.False(t, hasCycle)
	assert.Empty(t, cycle)
}

func TestDetectCycle_Diamond(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	graph := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {"d"},
	}

	// Act
	hasCycle, cycle := DetectCycle(graph, "a")

	// Assert
	assert.False(t, hasCycle)
	assert.Empty(t, cycle)
}

func TestDetectCycle_TwoNode(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	graph := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}

	// Act
	hasCycle, cycle := DetectCycle(graph, "a")

	// Assert
	assert.True(t, hasCycle)
	assert.Equal(t, []string{"a", "b", "a"}, cycle)
}

func TestDetectCycle_ThreeNode(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}

	// Act
	hasCycle, cycle := DetectCycle(graph, "a")

	// Assert
	assert.True(t, hasCycle)
	assert.Equal(t, []string{"a", "b", "c", "a"}, cycle)
}

func TestDetectCycle_CycleNotInvolvingStart(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"b"},
	}

	// Act
	hasCycle, cycle := DetectCycle(graph, "a")

	// Assert
	assert.True(t, hasCycle)
	assert.Equal(t, []string{"b", "c", "b"}, cycle)
}

func TestDetectCycle_StartNotInGraph(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	graph := map[string][]string{
		"x": {"y"},
	}

	// Act
	hasCycle, cycle := DetectCycle(graph, "a")

	// Assert
	assert.False(t, hasCycle)
	assert.Empty(t, cycle)
}
