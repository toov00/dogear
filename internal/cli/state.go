package cli

import (
	"context"
	"database/sql"

	"dogear/internal/repo"
)

type state struct {
	db    *sql.DB
	store *repo.Store
}

var appState *state

func store() *repo.Store {
	return appState.store
}

func dbCtx() context.Context {
	return context.Background()
}

func closeDB() error {
	if appState == nil || appState.db == nil {
		return nil
	}
	err := appState.db.Close()
	appState = nil
	return err
}
