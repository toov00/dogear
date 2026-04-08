package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func cmdRemove() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   `remove [flags] <title>`,
		Short: "Delete a title and every checkpoint.",
		Long:  "You must confirm in the terminal or pass --yes. This cannot be undone.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			t, err := resolveTitle(ctx, store(), args[0])
			if err != nil {
				return err
			}
			if !yes {
				fmt.Fprintf(os.Stderr, "Remove %q and all checkpoints? [y/N]: ", t.Title)
				sc := bufio.NewScanner(os.Stdin)
				if !sc.Scan() {
					return fmt.Errorf("cancelled")
				}
				line := strings.ToLower(strings.TrimSpace(sc.Text()))
				if line != "y" && line != "yes" {
					return fmt.Errorf("cancelled")
				}
			}
			return store().DeleteTitle(ctx, t.ID)
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "confirm deletion (required for non-interactive use)")
	return cmd
}
