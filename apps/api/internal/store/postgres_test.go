package store_test

import (
	"database/sql"
	"testing"

	"bookingms/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostgresReadinessAfterMigrations(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	const image = "postgres:17.11@sha256:ff87138949ad9a46f32ff23bd6852b52ce024a13dffe9b4087be98292cdd3b4f"
	ctx := t.Context()
	container, err := postgres.Run(ctx, image, postgres.BasicWaitStrategies())
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatal(err)
	}
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	ready, err := store.OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer ready.Close()
	if ready.Ready(ctx) {
		t.Fatal("la base recién creada todavía no tiene el esquema de Agenda")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpContext(ctx, db, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	if !ready.Ready(ctx) {
		t.Fatal("la base migrada debe estar disponible")
	}
}
