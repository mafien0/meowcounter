// Package db provides a counter store backed by sqlite (local) or Turso (libSQL).
package db

import (
	"context"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type Counter struct {
	Name  string
	Count int64
}

// DB is the counter store each backend implements.
type DB interface {
	GetCounter(ctx context.Context, name string) (Counter, error)
	UpsertCounter(ctx context.Context, name string, count int64) error
	Close() error
}

// Config selects a backend.
// Driver "sqlite" uses Path; driver "turso" uses URL and Token.
type Config struct {
	Driver string
	Path   string
	URL    string
	Token  string
}

// Open connects to the backend selected by cfg.Driver.
func Open(cfg Config) (DB, error) {
	switch cfg.Driver {
	case "sqlite":
		return openSQLite(cfg.Path)
	case "turso":
		return openTurso(cfg.URL, cfg.Token)
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q (want sqlite or turso)", cfg.Driver)
	}
}
