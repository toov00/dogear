package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"dogear/internal/output"
)

func cmdDoctor() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check for orphaned rows and other obvious issues.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			issues := store().Doctor(ctx)
			if len(issues) == 0 {
				output.DoctorOK(os.Stdout)
				return nil
			}
			for _, s := range issues {
				output.DoctorIssue(os.Stderr, "dogear: "+s)
			}
			return fmt.Errorf("database has %d issue(s); see messages above", len(issues))
		},
	}
}
