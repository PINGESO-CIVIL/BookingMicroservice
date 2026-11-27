package store

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres checks the availability of BookingMS's own database.
type Postgres struct {
	db *sql.DB
}

func OpenPostgres(dsn string) (*Postgres, error) {
	if dsn == "" {
		return &Postgres{}, nil
	}
	db, err := sql.Open("pgx", dsn)
	return &Postgres{db: db}, err
}

func (p *Postgres) Ready(ctx context.Context) bool {
	return p.db != nil && p.db.PingContext(ctx) == nil
}

func (p *Postgres) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}
