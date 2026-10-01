// Package db: sqlite3 db
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNotFound = errors.New("not found")

type DB struct {
	*sql.DB
}

type Counter struct {
	Name  string
	Count int64
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite3: %w", err)
	}
	conn.SetMaxOpenConns(1)
	return &DB{conn}, nil
}

func (db *DB) GetCounter(ctx context.Context, name string) (Counter, error) {
	var c Counter
	err := db.QueryRowContext(ctx, `SELECT name, count FROM counter WHERE name = ?`, name).Scan(&c.Name, &c.Count)
	if errors.Is(err, sql.ErrNoRows) {
		return Counter{name, 0}, ErrNotFound
	}
	return c, err
}

// UpsertCounter Inserts or Updates the counter
// I probably count do something like incrementCounter()
// But i think Upsert would be better,
// Because i already have a ready counter struct
func (db *DB) UpsertCounter(ctx context.Context, name string, count int64) error {
	_, err := db.ExecContext(ctx, `
	INSERT INTO counter (name, count) VALUES (?, ?)
	ON CONFLICT(name) DO UPDATE SET count = excluded.count
	`, name, count)
	return err
}
