# Load DB settings (POSTGRES_USER, POSTGRES_DB, ...) from .env
include .env
export

.PHONY: run db-seed

# Start the API server
run:
	go run cmd/main.go

# Insert seed data (cricket players) into the Postgres container.
# Requires the DB to be up: docker compose up -d
db-seed:
	docker compose exec -T database psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < seeds/001_players.sql
