-include .env

.PHONY: migrate-up migrate-down

migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is not set. Copy .env.example to .env and fill in your connection string." && exit 1)
	goose -dir sql/schema postgres "$(DATABASE_URL)" up

migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is not set. Copy .env.example to .env and fill in your connection string." && exit 1)
	goose -dir sql/schema postgres "$(DATABASE_URL)" down
