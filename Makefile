.PHONY: generate sqlc externos migrate

generate: sqlc externos
	pnpm codegen

sqlc:
	go -C apps/api run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

externos:
	go -C apps/api run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.2 -generate types,client -include-tags Notificaciones -package notificaciones -o internal/store/notificaciones/client.gen.go ../../contracts/externos/notificaciones.yaml

migrate:
	go -C apps/api run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir db/migrations -table public.goose_db_version up
