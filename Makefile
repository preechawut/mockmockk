.PHONY: run build test test-coverage vet lint fmt db-up migrate-up migrate-down migrate-create up down

run:
	go run ./cmd/api

build:
	@mkdir -p bin
	go build -o bin/api ./cmd/api

test:
	go test -race ./...

test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

vet:
	go vet ./...

lint:
	golangci-lint run

fmt:
	gofmt -w cmd internal

db-up:
	docker compose up -d --wait db

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate -path /migrations -database "postgres://app:localpassword@db:5432/app?sslmode=disable" down 1

migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=<migration_name>"; exit 1; fi
	bash ./scripts/create-migration.sh "$(NAME)"

up:
	docker compose up -d --build

down:
	docker compose down
