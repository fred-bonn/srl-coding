-include .env

.PHONY: migrate-up migrate-down seed

migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is not set. Copy .env.example to .env and fill in your connection string." && exit 1)
	goose -dir sql/schema postgres "$(DATABASE_URL)" up

migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is not set. Copy .env.example to .env and fill in your connection string." && exit 1)
	goose -dir sql/schema postgres "$(DATABASE_URL)" down

seed:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is not set. Copy .env.example to .env and fill in your connection string." && exit 1)
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed
