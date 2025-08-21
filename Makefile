SHELL := /bin/bash

.PHONY: run-api run-consumer tidy deps lint test migrate-install migrate-up migrate-down build up down

export GO111MODULE=on

run-api:
	go run ./cmd/api

run-consumer:
	go run ./cmd/consumer

build:
	go build -o bin/api ./cmd/api
	go build -o bin/consumer ./cmd/consumer

tidy:
	go mod tidy

deps:
	go get \
			github.com/gin-gonic/gin \
			github.com/segmentio/kafka-go \
			github.com/jackc/pgx/v5/pgxpool \
			go.uber.org/zap \
			github.com/kelseyhightower/envconfig

migrate-up:
	@migrate -path ./internal/migrations -database "$$PG_DSN" up

migrate-down:
	@migrate -path ./internal/migrations -database "$$PG_DSN" down 1
