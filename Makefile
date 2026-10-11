.PHONY: run migrate test test-coverage build fmt vet lint db up down

run:
	go run ./cmd/api

migrate:
	go run ./cmd/migrate

test:
	go test -race ./...

test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

build:
	@mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate

fmt:
	gofmt -w cmd internal migrations

vet:
	go vet ./...

lint:
	golangci-lint run

db:
	docker compose up -d --wait db

up:
	docker compose up -d --build

down:
	docker compose down
