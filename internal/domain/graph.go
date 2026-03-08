package domain

import (
	"time"

	"github.com/w-h-a/bees/internal/util/dfs"
)

type Node struct {
	ID           string
	Title        string
	Status       Status
	Priority     int
	Type         Type
	DeferUntil   *time.Time
	EstimateMins int
}

type Edge struct {
	From string
	To   string
}

type Graph struct {
	Nodes map[string]Node
	Edges []Edge
}

func (g Graph) Subgraph(rootID string) Graph {
	adj := map[string][]string{}
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
		adj[e.To] = append(adj[e.To], e.From)
	}

	reachable := dfs.Reachable(adj, rootID)

	nodes := make(map[string]Node)
	var edges []Edge

	for id := range reachable {
		if n, ok := g.Nodes[id]; ok {
			nodes[id] = n
		}
	}

	for _, e := range g.Edges {
		if reachable[e.From] && reachable[e.To] {
			edges = append(edges, e)
		}
	}

	return Graph{Nodes: nodes, Edges: edges}
}

func (g Graph) FilterByStatus(status string) Graph {
	if status == "all" {
		return g
	}

	nodes := make(map[string]Node, len(g.Nodes))
	for id, n := range g.Nodes {
		switch status {
		case "":
			if n.Status == StatusClosed {
				continue
			}
		default:
			if string(n.Status) != status {
				continue
			}
		}
		nodes[id] = n
	}

	edges := make([]Edge, 0, len(g.Edges))
	for _, e := range g.Edges {
		if _, fromOk := nodes[e.From]; !fromOk {
			continue
		}
		if _, toOk := nodes[e.To]; !toOk {
			continue
		}
		edges = append(edges, e)
	}

	return Graph{Nodes: nodes, Edges: edges}
}

func BuildGraph(deps []Dependency, issues map[string]Issue) Graph {
	nodes := map[string]Node{}
	edges := []Edge{}

	for _, dep := range deps {
		fromIss, fromOk := issues[dep.DependsOnID]
		toIss, toOk := issues[dep.IssueID]

		if !fromOk || !toOk {
			continue
		}

		edges = append(edges, Edge{
			From: dep.DependsOnID,
			To:   dep.IssueID,
		})

		for _, entry := range []struct {
			id  string
			iss Issue
		}{
			{dep.DependsOnID, fromIss},
			{dep.IssueID, toIss},
		} {
			if _, ok := nodes[entry.id]; !ok {
				pri := 2
				if entry.iss.Priority != nil {
					pri = *entry.iss.Priority
				}
				nodes[entry.id] = Node{
					ID:           entry.id,
					Title:        entry.iss.Title,
					Status:       entry.iss.Status,
					Priority:     pri,
					Type:         entry.iss.Type,
					DeferUntil:   entry.iss.DeferUntil,
					EstimateMins: entry.iss.EstimateMins,
				}
			}
		}
	}

	return Graph{Nodes: nodes, Edges: edges}
}
