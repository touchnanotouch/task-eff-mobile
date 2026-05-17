.PHONY: up down restart logs clean shell

COMPOSE_FILE = docker-compose.yml

help:
	@echo "Commands: swag, up, down, restart, logs, clean, shell, db, migrate-up, migrate-down, test"

swag:
	swag init -g cmd/server/main.go

up:
	docker compose -f $(COMPOSE_FILE) up -d

down:
	docker compose -f $(COMPOSE_FILE) down

build:
	docker compose -f $(COMPOSE_FILE) build --no-cache

restart:
	docker compose -f $(COMPOSE_FILE) restart

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

clean:
	docker compose -f $(COMPOSE_FILE) down -v

shell:
	docker compose -f $(COMPOSE_FILE) exec app sh

db:
	docker compose -f $(COMPOSE_FILE) exec postgres psql -U postgres -d subscriptions

migrate-up:
	docker compose -f $(COMPOSE_FILE) exec -T postgres psql -U postgres -d subscriptions < migrations/001_create_subscriptions.up.sql

migrate-down:
	docker compose -f $(COMPOSE_FILE) exec -T postgres psql -U postgres -d subscriptions < migrations/001_create_subscriptions.down.sql

test:
	go test ./...
