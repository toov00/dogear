package cli

import (
	"time"

	"github.com/spf13/cobra"
)

func cmdFinish() *cobra.Command {
	return &cobra.Command{
		Use:   `finish <title>`,
		Short: "Mark a title finished (checkpoints stay).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			t, err := resolveTitle(ctx, store(), args[0])
			if err != nil {
				return err
			}
			return store().Finish(ctx, t.ID, time.Now())
		},
	}
}
