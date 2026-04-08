package cli

import (
	"context"
	"fmt"
	"os"

	"dogear/internal/model"
	"dogear/internal/output"
	"dogear/internal/repo"
	"dogear/internal/search"
)

func resolveTitle(ctx context.Context, st *repo.Store, query string) (*model.Title, error) {
	all, err := st.LoadAllTitles(ctx)
	if err != nil {
		return nil, err
	}
	t, err := search.BestSingleTitle(query, all)
	if err != nil {
		if amb, ok := search.IsAmbiguous(err); ok {
			output.Ambiguous(os.Stderr, *amb)
			return nil, fmt.Errorf("ambiguous title")
		}
		if search.IsNoMatch(err) {
			return nil, fmt.Errorf("no matching title for %q", query)
		}
		return nil, err
	}
	return st.LoadTitleFull(ctx, t.ID)
}
