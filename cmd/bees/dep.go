package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/w-h-a/bees/internal/domain"
)

func newDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Manage and graph issue dependencies",
	}

	cmd.AddCommand(newDepAddCmd())
	cmd.AddCommand(newDepRemoveCmd())
	cmd.AddCommand(newDepGraphCmd())

	return cmd
}

func newDepAddCmd() *cobra.Command {
	var blockedID string

	cmd := &cobra.Command{
		Use:   "add <blocker-id> --blocks <blocked-id>",
		Short: "Add a blocking dependency",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			blocker, blocked, err := svc.AddDependency(cmd.Context(), args[0], blockedID)
			if err != nil {
				return err
			}

			if !jsonOutput {
				fmt.Printf("%s now blocks %s\n", blocker, blocked)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(map[string]string{
				"blocker_id": blocker,
				"blocked_id": blocked,
				"action":     "added",
			})
		},
	}

	cmd.Flags().StringVar(&blockedID, "blocks", "", "ID of the issue being blocked")
	cmd.MarkFlagRequired("blocks")

	return cmd
}

func newDepRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <blocker-id> <blocked-id>",
		Short: "Remove a blocking dependency",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			blocker, blocked, changed, err := svc.RemoveDependency(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}

			if !jsonOutput {
				if !changed {
					fmt.Printf("No dependency: %s does not block %s\n", blocker, blocked)
				} else {
					fmt.Printf("%s no longer blocks %s\n", blocker, blocked)
				}
				return nil
			}

			action := "removed"
			if !changed {
				action = "none"
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(map[string]string{
				"blocker_id": blocker,
				"blocked_id": blocked,
				"action":     action,
			})
		},
	}

	return cmd
}

func newDepGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph [<id>]",
		Short: "Show the dependency graph",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id *string
			if len(args) == 1 {
				id = &args[0]
			}

			graph, err := svc.BuildGraph(cmd.Context(), id)
			if err != nil {
				return err
			}

			for id, n := range graph.Nodes {
				if n.Status == domain.StatusClosed {
					delete(graph.Nodes, id)
				}
			}
			filtered := make([]domain.Edge, 0, len(graph.Edges))
			for _, e := range graph.Edges {
				if _, fromOk := graph.Nodes[e.From]; !fromOk {
					continue
				}
				if _, toOk := graph.Nodes[e.To]; !toOk {
					continue
				}
				filtered = append(filtered, e)
			}
			graph.Edges = filtered

			if !jsonOutput {
				printGraph(graph)
				return nil
			}

			type jsonNode struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Status   string `json:"status"`
				Priority int    `json:"priority"`
			}
			type jsonEdge struct {
				From string `json:"from"`
				To   string `json:"to"`
			}
			type jsonGraph struct {
				Nodes []jsonNode `json:"nodes"`
				Edges []jsonEdge `json:"edges"`
			}

			nodes := make([]jsonNode, 0, len(graph.Nodes))
			for _, n := range graph.Nodes {
				nodes = append(nodes, jsonNode{
					ID:       n.ID,
					Title:    n.Title,
					Status:   string(n.Status),
					Priority: n.Priority,
				})
			}
			sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })

			edges := make([]jsonEdge, 0, len(graph.Edges))
			for _, e := range graph.Edges {
				edges = append(edges, jsonEdge{From: e.From, To: e.To})
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(jsonGraph{Nodes: nodes, Edges: edges})
		},
	}

	return cmd
}
