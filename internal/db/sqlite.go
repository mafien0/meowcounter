package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type sqliteDB struct {
	*sql.DB
}

func openSQLite(path string) (DB, error) {
	if path == "" {
		return nil, errors.New("SQLITE_PATH is not set")
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1)
	return &sqliteDB{conn}, nil
}

func (db *sqliteDB) GetCounter(ctx context.Context, name string) (Counter, error) {
	var c Counter
	err := db.QueryRowContext(ctx, `SELECT name, count FROM counter WHERE name = ?`, name).Scan(&c.Name, &c.Count)
	if errors.Is(err, sql.ErrNoRows) {
		return Counter{Name: name}, ErrNotFound
	}
	return c, err
}

func (db *sqliteDB) UpsertCounter(ctx context.Context, name string, count int64) error {
	_, err := db.ExecContext(ctx, `
	INSERT INTO counter (name, count) VALUES (?, ?)
	ON CONFLICT(name) DO UPDATE SET count = excluded.count
	`, name, count)
	return err
}
