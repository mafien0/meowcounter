package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type tursoDB struct {
	*sql.DB
}

func openTurso(dbURL, authToken string) (DB, error) {
	if dbURL == "" {
		return nil, errors.New("TURSO_DATABASE_URL is not set")
	}
	if authToken == "" {
		return nil, errors.New("TURSO_AUTH_TOKEN is not set")
	}

	u, err := url.Parse(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse TURSO_DATABASE_URL: %w", err)
	}
	q := u.Query()
	q.Set("authToken", authToken)
	u.RawQuery = q.Encode()

	conn, err := sql.Open("libsql", u.String())
	if err != nil {
		return nil, fmt.Errorf("open libsql: %w", err)
	}
	conn.SetMaxOpenConns(1)
	return &tursoDB{conn}, nil
}

func (db *tursoDB) GetCounter(ctx context.Context, name string) (Counter, error) {
	var c Counter
	err := db.QueryRowContext(ctx, `SELECT name, count FROM counter WHERE name = ?`, name).Scan(&c.Name, &c.Count)
	if errors.Is(err, sql.ErrNoRows) {
		return Counter{Name: name}, ErrNotFound
	}
	return c, err
}

func (db *tursoDB) UpsertCounter(ctx context.Context, name string, count int64) error {
	_, err := db.ExecContext(ctx, `
	INSERT INTO counter (name, count) VALUES (?, ?)
	ON CONFLICT(name) DO UPDATE SET count = excluded.count
	`, name, count)
	return err
}
