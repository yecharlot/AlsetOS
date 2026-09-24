#!/bin/sh
set -eu

echo "[1/6] módulos"
go mod tidy

echo "[2/6] suite"
go test ./...

echo "[3/6] organismo"
go run ./cmd/alsetos organismos/organismo-si.alset

echo "[4/6] sandbox"
go run ./cmd/sandbox

echo "[5/6] autonomía"
go run ./cmd/alset-autonomia --duracion 3s

echo "[6/6] kernel"
set +e
timeout 5s go run ./cmd/alset-kernel
estado=$?
set -e
if [ "$estado" -ne 124 ]; then
  echo "kernel terminó con código inesperado: $estado"
  exit "$estado"
fi

echo "[ALSET-FINAL] suite completada"
