package cli

import (
	"os"

	"github.com/spf13/cobra"

	"dogear/internal/output"
)

func cmdWhere() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     `where [flags] <title>`,
		Aliases: []string{"current"},
		Short:   "Print the latest checkpoint for a title (alias: current).",
		Long:    "Resolves <title> with fuzzy matching when there is no exact match. Use list or search if several items are similar.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			t, err := resolveTitle(ctx, store(), args[0])
			if err != nil {
				return err
			}
			if asJSON {
				return output.WriteWhereJSON(os.Stdout, t)
			}
			output.Where(os.Stdout, t)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON instead of text")
	return cmd
}
