package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"dogear/internal/model"
	"dogear/internal/position"
)

func cmdAdd() *cobra.Command {
	var (
		author        string
		format        string
		page          int
		chapter       int
		totalPages    int
		totalChapters int
		section       string
		loc           int
		percent       float64
		note          string
		tags          []string
		allowDup      bool
	)
	cmd := &cobra.Command{
		Use:   `add [flags] <title>`,
		Short: "Start tracking a new title.",
		Long:  "Optionally record an initial checkpoint with --page, --section, --loc, etc. Use repeated --tag or commas for several tags.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.TrimSpace(args[0])
			if title == "" {
				return fmt.Errorf("title is empty")
			}
			ctx := dbCtx()
			dup, err := store().HasDuplicateActive(ctx, title, format)
			if err != nil {
				return err
			}
			if dup && !allowDup {
				return fmt.Errorf("active title with same name and format exists (use --allow-duplicate to override)")
			}
			in := model.TitleInput{
				Title:          title,
				AuthorOrSource: author,
				Format:         format,
			}
			if cmd.Flags().Changed("total-pages") {
				in.TotalPages = &totalPages
			}
			if cmd.Flags().Changed("total-chapters") {
				in.TotalChapters = &totalChapters
			}
			tagList := expandTags(tags)
			now := time.Now()
			id, err := store().InsertTitle(ctx, in, now)
			if err != nil {
				return err
			}
			if len(tagList) > 0 {
				if err := store().ReplaceTags(ctx, id, tagList); err != nil {
					return err
				}
			}
			var pageP *int
			if cmd.Flags().Changed("page") {
				pageP = &page
			}
			var chapterP *int
			if cmd.Flags().Changed("chapter") {
				chapterP = &chapter
			}
			var locP *int
			if cmd.Flags().Changed("loc") {
				locP = &loc
			}
			var secP *string
			if strings.TrimSpace(section) != "" {
				s := strings.TrimSpace(section)
				secP = &s
			}
			var noteP *string
			if strings.TrimSpace(note) != "" {
				n := strings.TrimSpace(note)
				noteP = &n
			}
			var pctP *float64
			if cmd.Flags().Changed("percent") {
				pctP = &percent
			}
			cpIn, err := position.BuildOptional(pageP, chapterP, secP, locP, pctP, noteP)
			if err != nil {
				return err
			}
			if cpIn != nil {
				if _, err := store().InsertCheckpoint(ctx, id, *cpIn, now); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&author, "author", "", "author, editor, or source name")
	cmd.Flags().StringVar(&format, "format", "", "binding or medium (paperback, pdf, epub, ...)")
	cmd.Flags().IntVar(&page, "page", 0, "initial page")
	cmd.Flags().IntVar(&chapter, "chapter", 0, "initial chapter")
	cmd.Flags().IntVar(&totalPages, "total-pages", 0, "total pages (for progress)")
	cmd.Flags().IntVar(&totalChapters, "total-chapters", 0, "total chapters")
	cmd.Flags().StringVar(&section, "section", "", "initial section label")
	cmd.Flags().IntVar(&loc, "loc", 0, "initial location / offset")
	cmd.Flags().Float64Var(&percent, "percent", 0, "initial percent (0-100)")
	cmd.Flags().StringVar(&note, "note", "", "note on first checkpoint")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tag(s); repeat flag or use commas")
	cmd.Flags().BoolVar(&allowDup, "allow-duplicate", false, "allow another active row with same title and format")
	return cmd
}

func expandTags(ss []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, s := range ss {
		for _, p := range strings.Split(s, ",") {
			p = strings.TrimSpace(strings.ToLower(p))
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}
