package cli

import (
	"os"

	"github.com/spf13/cobra"

	"dogear/internal/output"
)

func cmdList() *cobra.Command {
	var active, finished bool
	var tag string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List titles (active first, then by last update).",
		Long:  "Default lists everything. Narrow with --active, --finished, or --tag. Position tallies use each title's saved metadata.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			st := ""
			if active && !finished {
				st = "active"
			} else if finished && !active {
				st = "finished"
			}
			list, err := store().ListTitles(ctx, st, tag)
			if err != nil {
				return err
			}
			if asJSON {
				return output.WriteListJSON(os.Stdout, list)
			}
			output.ListTable(os.Stdout, list)
			return nil
		},
	}
	cmd.Flags().BoolVar(&active, "active", false, "only active titles")
	cmd.Flags().BoolVar(&finished, "finished", false, "only finished titles")
	cmd.Flags().StringVar(&tag, "tag", "", "filter by tag (one tag; repeat list for several filters)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON array instead of a table")
	return cmd
}
