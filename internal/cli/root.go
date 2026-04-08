package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"dogear/internal/db"
	"dogear/internal/repo"
)

var flagDB string

func Execute() int {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		if msg := err.Error(); msg != "" && !strings.HasPrefix(msg, "dogear:") {
			fmt.Fprintf(os.Stderr, "dogear: %s\n", msg)
		} else if msg != "" {
			fmt.Fprintf(os.Stderr, "%s\n", msg)
		}
		_ = closeDB()
		return 1
	}
	if err := closeDB(); err != nil {
		fmt.Fprintf(os.Stderr, "dogear: close database: %v\n", err)
		return 1
	}
	return 0
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:              "dogear",
		Short:            "Remember where you left off in books, papers, and notes.",
		Long:             "dogear stores reading positions as checkpoints in a local SQLite file. No accounts, no cloud.",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			path, err := db.ResolvePath(flagDB)
			if err != nil {
				return fmt.Errorf("resolve database path: %w", err)
			}
			conn, err := db.Open(path)
			if err != nil {
				return fmt.Errorf("open database %s: %w", path, err)
			}
			appState = &state{db: conn, store: repo.New(conn)}
			return nil
		},
	}
	root.PersistentFlags().StringVar(&flagDB, "db", "", `SQLite database path (overrides $DOGEAR_DB; default: config dir / "dogear" / "dogear.db")`)
	root.AddCommand(
		cmdAdd(),
		cmdUpdate(),
		cmdWhere(),
		cmdList(),
		cmdLately(),
		cmdHistory(),
		cmdFinish(),
		cmdRemove(),
		cmdStats(),
		cmdSearch(),
		cmdTag(),
		cmdUntag(),
		cmdStale(),
		cmdExport(),
		cmdImport(),
		cmdDoctor(),
	)
	return root
}
