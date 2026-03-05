package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of bees.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			v, c, d := "dev", "unknown", "unknown"

			if info, ok := debug.ReadBuildInfo(); ok {
				if info.Main.Version != "" && info.Main.Version != "(devel)" {
					v = info.Main.Version
				}
				for _, s := range info.Settings {
					switch s.Key {
					case "vcs.revision":
						c = s.Value
					case "vcs.time":
						d = s.Value
					}
				}
			}

			if !jsonOutput {
				short := c
				if len(short) > 7 {
					short = short[:7]
				}
				fmt.Printf("bees %s (commit: %s, built: %s)\n", v, short, d)
				return nil
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", " ")

			return enc.Encode(map[string]string{
				"version": v,
				"commit":  c,
				"date":    d,
			})
		},
	}

	return cmd
}
