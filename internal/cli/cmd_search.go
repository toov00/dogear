package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"dogear/internal/output"
	"dogear/internal/search"
)

func cmdSearch() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   `search [flags] <query>`,
		Short: "Fuzzy-match titles by name.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			all, err := store().LoadAllTitles(ctx)
			if err != nil {
				return err
			}
			matches := search.MatchTitles(strings.TrimSpace(args[0]), all, 0)
			if len(matches) == 0 {
				return fmt.Errorf("no matches for %q", args[0])
			}
			if asJSON {
				return output.WriteSearchJSON(os.Stdout, matches)
			}
			output.SearchResults(os.Stdout, matches)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON instead of plain lines")
	return cmd
}
