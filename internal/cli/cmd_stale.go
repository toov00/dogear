package cli

import (
	"os"

	"github.com/spf13/cobra"

	"dogear/internal/output"
)

func cmdStale() *cobra.Command {
	var days int
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "stale [flags]",
		Short: "Active titles with no update for at least N days (default 30).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			list, err := store().StaleActive(ctx, days)
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
	cmd.Flags().IntVar(&days, "days", 30, "minimum days since last update")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON array instead of a table")
	return cmd
}
