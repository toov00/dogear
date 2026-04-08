package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func cmdTag() *cobra.Command {
	return &cobra.Command{
		Use:   `tag <title> <tag>`,
		Short: "Attach a tag to a title (add more with another tag run).",
		Long:  "The tag is one argument; put multi-word tags in quotes. Title may be several words: everything before the last argument is the title.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			titleq := strings.Join(args[:len(args)-1], " ")
			tagVal := args[len(args)-1]
			t, err := resolveTitle(ctx, store(), titleq)
			if err != nil {
				return err
			}
			return store().AddTag(ctx, t.ID, tagVal)
		},
	}
}

func cmdUntag() *cobra.Command {
	return &cobra.Command{
		Use:   `untag <title> <tag>`,
		Short: "Remove one tag from a title.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := dbCtx()
			titleq := strings.Join(args[:len(args)-1], " ")
			tagVal := args[len(args)-1]
			t, err := resolveTitle(ctx, store(), titleq)
			if err != nil {
				return err
			}
			return store().RemoveTag(ctx, t.ID, tagVal)
		},
	}
}
