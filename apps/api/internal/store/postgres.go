package store

import (
	"context"
	"database/sql"

	"bookingms/internal/repository"

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
	if p.db == nil {
		return false
	}
	ready, err := repository.New(p.db).DatabaseReady(ctx)
	return err == nil && ready
}

func (p *Postgres) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}
