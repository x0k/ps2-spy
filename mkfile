#!/bin/bash -e

env:
  if [ -f .env.local ]
  then
    export $(cat .env.local | xargs)
  fi
  if [ -f ~/Sync/ps2-spy/.env ]
  then
    export $(cat ~/Sync/ps2-spy/.env | xargs)
  fi

# Dev
d: env
  go run cmd/ps2-spy/main.go --config config/dev.yml

# Build bin
b:
  go build -o bin/app cmd/ps2-spy/main.go

t:
  go test ./...

## DATABASE

db:
  sqlc generate

qlint:
  sqlc vet

migration: env
  migrate create -ext sql -dir db/migrations -seq $1

migrate-down:
  migrate -source file://db/migrations -database sqlite3://storage/storage.db down 1

text:
  go generate ./internal/translations/translations.go

p: env
  .bin/app --config config/dev.yml

prof: env
  # heap, goroutine, block, threadcreate, mutex
  go tool pprof -http :8081 "${PPROF_ADDRESS}/debug/pprof/$1"

test:
  go test ./...

lint:
  golangci-lint run ./...

up:
  USER_ID="$(id -u)" docker-compose -f monitoring/docker-compose.yml up -d

down:
  docker-compose -f monitoring/docker-compose.yml down

o:
  go run tools/explorer/*.go --resource outfit --query $1 --output tmp

os:
  go run tools/explorer/*.go --resource outfits --query $1 --output tmp

c:
  go run tools/explorer/*.go --resource character --query $1 --output tmp

cs:
  go run tools/explorer/*.go --resource characters --query $1 --output tmp

ms:
  go run tools/explorer/*.go --resource members --query $1 --output tmp

ws:
  go run tools/explorer/*.go --resource world --query $1 --output tmp
