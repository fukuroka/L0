SHELL := /bin/bash

.PHONY: run-api run-consumer tidy deps lint test migrate-install migrate-up migrate-down build up down

export GO111MODULE=on

run:
	go run ./cmd/main.go

build:
	go build -o bin/main ./cmd/main.go


tidy:
	go mod tidy

deps:
	go get \
		github.com/gin-gonic/gin \
		github.com/jackc/pgx/v5 \
		github.com/joho/godotenv \
		github.com/kelseyhightower/envconfig \
		github.com/segmentio/kafka-go \
		github.com/swaggo/files \
		github.com/swaggo/gin-swagger \
		github.com/swaggo/swag \

migrate-up:
	@migrate -path ./internal/migrations -database "$$PG_DSN" up

migrate-down:
	@migrate -path ./internal/migrations -database "$$PG_DSN" down 1

up:
	@docker-compose up -d

down:
	@docker-compose down
