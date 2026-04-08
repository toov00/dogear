package cli

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func cmdExport() *cobra.Command {
	var outPath string
	cmd := &cobra.Command{
		Use:   "export [flags]",
		Short: "Dump the whole library as JSON to stdout or --out.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			doc, err := store().ExportSnapshot(ctx)
			if err != nil {
				return err
			}
			b, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				return err
			}
			if outPath != "" {
				return os.WriteFile(outPath, append(b, '\n'), 0o644)
			}
			if _, err := os.Stdout.Write(append(b, '\n')); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "write to this file instead of stdout")
	return cmd
}
