package cli

import (
	"os"

	"github.com/spf13/cobra"

	"dogear/internal/output"
)

func cmdHistory() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   `history [flags] <title>`,
		Short: "List checkpoints for a title, newest first.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			t, err := resolveTitle(ctx, store(), args[0])
			if err != nil {
				return err
			}
			cps, err := store().History(ctx, t.ID)
			if err != nil {
				return err
			}
			if asJSON {
				return output.WriteHistoryJSON(os.Stdout, t, cps)
			}
			output.HistoryHeader(os.Stdout, t)
			for _, cp := range cps {
				output.HistoryLine(os.Stdout, cp, t)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON instead of text")
	return cmd
}
