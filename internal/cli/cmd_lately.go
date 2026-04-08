package cli

import (
	"os"

	"github.com/spf13/cobra"

	"dogear/internal/output"
)

func cmdLately() *cobra.Command {
	var n int
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "lately [flags]",
		Short: "Active titles sorted by most recent update.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			list, err := store().Lately(ctx, n)
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
	cmd.Flags().IntVar(&n, "n", 20, "maximum titles")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON array instead of a table")
	return cmd
}
