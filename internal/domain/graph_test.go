package domain

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildGraph_DisconnectedComponents(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	p1 := 1
	p3 := 3
	defer1 := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

	issues := map[string]Issue{
		"a": {ID: "a", Title: "Alpha", Status: StatusOpen, Type: TypeTask, Priority: &p1, DeferUntil: &defer1, EstimateMins: 30},
		"b": {ID: "b", Title: "Beta", Status: StatusInProgress, Type: TypeBug, Priority: &p3, EstimateMins: 60},
		"c": {ID: "c", Title: "Charlie", Status: StatusOpen, Type: TypeFeature, Priority: &p1, DeferUntil: &defer1},
		"d": {ID: "d", Title: "Delta", Status: StatusClosed, Type: TypeChore, Priority: &p3},
	}

	deps := []Dependency{
		{DependsOnID: "a", IssueID: "b"},
		{DependsOnID: "c", IssueID: "d"},
	}

	// Act
	g := BuildGraph(deps, issues)

	// Assert
	require.Len(t, g.Nodes, 4)
	require.Len(t, g.Edges, 2)

	require.Equal(t, Node{ID: "a", Title: "Alpha", Status: StatusOpen, Priority: 1, Type: TypeTask, DeferUntil: &defer1, EstimateMins: 30}, g.Nodes["a"])
	require.Equal(t, Node{ID: "b", Title: "Beta", Status: StatusInProgress, Priority: 3, Type: TypeBug, EstimateMins: 60}, g.Nodes["b"])
	require.Equal(t, Node{ID: "c", Title: "Charlie", Status: StatusOpen, Priority: 1, Type: TypeFeature, DeferUntil: &defer1}, g.Nodes["c"])
	require.Equal(t, Node{ID: "d", Title: "Delta", Status: StatusClosed, Priority: 3, Type: TypeChore}, g.Nodes["d"])

	require.Contains(t, g.Edges, Edge{From: "a", To: "b"})
	require.Contains(t, g.Edges, Edge{From: "c", To: "d"})
}

func TestBuildGraph_SubgraphFromRoot(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	p2 := 2

	issues := map[string]Issue{
		"a": {ID: "a", Title: "Alpha", Status: StatusOpen, Priority: &p2},
		"b": {ID: "b", Title: "Beta", Status: StatusOpen, Priority: &p2},
		"c": {ID: "c", Title: "Charlie", Status: StatusOpen, Priority: &p2},
		"d": {ID: "d", Title: "Delta", Status: StatusOpen, Priority: &p2},
		"e": {ID: "e", Title: "Echo", Status: StatusOpen, Priority: &p2},
	}

	// a blocks b, b blocks c (one component); d blocks e (disconnected component)
	deps := []Dependency{
		{DependsOnID: "a", IssueID: "b"},
		{DependsOnID: "b", IssueID: "c"},
		{DependsOnID: "d", IssueID: "e"},
	}

	full := BuildGraph(deps, issues)

	// Act
	sub := full.Subgraph("a")

	// Assert
	require.Len(t, sub.Nodes, 3)
	require.Contains(t, sub.Nodes, "a")
	require.Contains(t, sub.Nodes, "b")
	require.Contains(t, sub.Nodes, "c")
	require.NotContains(t, sub.Nodes, "d")
	require.NotContains(t, sub.Nodes, "e")

	require.Len(t, sub.Edges, 2)
	require.Contains(t, sub.Edges, Edge{From: "a", To: "b"})
	require.Contains(t, sub.Edges, Edge{From: "b", To: "c"})
}

func TestBuildGraph_EmptyInput(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Act
	g := BuildGraph(nil, nil)

	// Assert
	require.Empty(t, g.Nodes)
	require.Empty(t, g.Edges)
}
