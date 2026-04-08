package cli

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"dogear/internal/output"
	"dogear/internal/position"
)

func cmdUpdate() *cobra.Command {
	var (
		page    int
		chapter int
		section string
		loc     int
		percent float64
		note    string
		tags    []string
		asJSON  bool
	)
	cmd := &cobra.Command{
		Use:   `update [flags] <title>`,
		Short: "Append a checkpoint (and optionally more tags).",
		Long:  "At least one position field or --note is required. Tags are merged; repeat --tag or use commas.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			t, err := resolveTitle(ctx, store(), args[0])
			if err != nil {
				return err
			}
			var pageP *int
			if cmd.Flags().Changed("page") {
				pageP = &page
			}
			var chP *int
			if cmd.Flags().Changed("chapter") {
				chP = &chapter
			}
			var secP *string
			if strings.TrimSpace(section) != "" {
				s := strings.TrimSpace(section)
				secP = &s
			}
			var locP *int
			if cmd.Flags().Changed("loc") {
				locP = &loc
			}
			var pctP *float64
			if cmd.Flags().Changed("percent") {
				pctP = &percent
			}
			var noteP *string
			if strings.TrimSpace(note) != "" {
				n := strings.TrimSpace(note)
				noteP = &n
			}
			in, err := position.BuildInput(pageP, chP, secP, locP, pctP, noteP)
			if err != nil {
				return err
			}
			now := time.Now()
			if _, err := store().InsertCheckpoint(ctx, t.ID, in, now); err != nil {
				return err
			}
			for _, tg := range expandTags(tags) {
				if err := store().AddTag(ctx, t.ID, tg); err != nil {
					return err
				}
			}
			full, err := store().LoadTitleFull(ctx, t.ID)
			if err != nil {
				return err
			}
			if asJSON {
				return output.WriteWhereJSON(os.Stdout, full)
			}
			output.Where(os.Stdout, full)
			return nil
		},
	}
	cmd.Flags().IntVar(&page, "page", 0, "page number")
	cmd.Flags().IntVar(&chapter, "chapter", 0, "chapter number")
	cmd.Flags().StringVar(&section, "section", "", "section label (e.g. \"3.4\")")
	cmd.Flags().IntVar(&loc, "loc", 0, "e-reader style location / offset")
	cmd.Flags().Float64Var(&percent, "percent", 0, "percent complete (0-100)")
	cmd.Flags().StringVar(&note, "note", "", "free-form note")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "add tag(s); repeat or comma-separated")
	cmd.Flags().BoolVar(&asJSON, "json", false, "after update, print the title as JSON")
	return cmd
}
