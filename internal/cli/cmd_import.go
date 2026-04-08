package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"dogear/internal/model"
)

func cmdImport() *cobra.Command {
	var replace bool
	cmd := &cobra.Command{
		Use:   `import [flags] <file>`,
		Short: "Load a JSON export (replaces the DB only with --replace).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !replace {
				return fmt.Errorf("refusing import without --replace (this wipes existing data)")
			}
			b, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			var doc model.ExportDoc
			if err := json.Unmarshal(b, &doc); err != nil {
				return fmt.Errorf("parse JSON: %w", err)
			}
			ctx := dbCtx()
			return store().ImportReplace(ctx, &doc)
		},
	}
	cmd.Flags().BoolVar(&replace, "replace", false, "delete local data and load this file (required)")
	return cmd
}
