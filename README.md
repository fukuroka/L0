# Сервис заказов

Простой сервис на Go с HTTP API, Kafka (producer/consumer) и хранением в Postgres.

## Требования
- Go 1.24+
- Docker и Docker Compose (для локального Postgres и Kafka)

## Быстрый старт

1. Скопируйте пример переменных окружения и при необходимости отредактируйте:

```bash
cp .env.example .env
```

2. Запустите инфраструктуру:

```bash
docker-compose up -d
```

3. Установите зависимости и соберите проект:

```bash
make deps
make build
```

4. Запустите API:

```bash
make run-api
```

5. Доступные эндпоинты:
- `GET /healthcheck`
- `GET /orders/` (опциональный query: `limit`)
- `GET /orders/:id`
- `POST /orders/` — сгенерировать случайный заказ и опубликовать в Kafka

## Переменные окружения
Все переменные перечислены в файле `.env.example`.

## Make цели
- `make deps` — скачать зависимости
- `make build` — собрать бинарники `api` и `consumer`
- `make run-api` — запустить API
- `make run-consumer` — запустить consumer
- `make migrate-up` / `make migrate-down` — выполнить миграции БД

## Примечания
- Для локального запуска скопируйте `.env.example` в `.env`: `cp .env.example .env`

