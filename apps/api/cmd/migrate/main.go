package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := migrate(); err != nil {
		log.Fatal(err)
	}
	log.Print("Migraciones de BookingMS al día")
}

func migrate() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL es requerida para migrar")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return errors.New("configuración inválida de PostgreSQL")
	}
	defer db.Close()

	// Creating the schema named after the PostgreSQL role changes its current schema.
	// Keep the migration history in the same schema before and after that change.
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS("/migrations"),
		goose.WithTableName("public.goose_db_version"))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err = provider.Up(ctx)
	return err
}
