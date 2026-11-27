-- name: DatabaseReady :one
SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname = 'bookingms') AS ready;
