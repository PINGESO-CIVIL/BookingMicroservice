#!/bin/sh
set -eu
cd "$(dirname "$0")"
pendientes=$(gofmt -l .)
if [ -n "$pendientes" ]; then
  printf 'Archivos pendientes de gofmt:\n%s\n' "$pendientes" >&2
  exit 1
fi
go vet ./...
