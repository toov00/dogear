package cli

import (
	"os"

	"github.com/spf13/cobra"

	"dogear/internal/output"
)

func cmdStats() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "stats [flags]",
		Short: "Lightweight counts and cues about your library.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			active, finished, cps, err := store().Counts(ctx)
			if err != nil {
				return err
			}
			recent, err := store().MostRecentlyUpdatedTitle(ctx)
			if err != nil {
				return err
			}
			stale, err := store().StalestActive(ctx)
			if err != nil {
				return err
			}
			avg, err := store().PageStepAverages(ctx)
			if err != nil {
				return err
			}
			if asJSON {
				return output.WriteStatsJSON(os.Stdout, active, finished, cps, recent, stale, avg)
			}
			output.Stats(os.Stdout, active, finished, cps, recent, stale, avg)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON instead of text")
	return cmd
}
