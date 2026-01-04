.SILENT:
.PHONY: run sql migrate-up migrate-down migrate-sync mock-migrate-up mock-migrate-down templ-generate templ-watch tailwind-build tailwind-watch build generate

# Run the bot
run:
	go run ./cmd/betrayal-bot/main.go

# Run web server only (no Discord bot)
run-web:
	DISABLE_DISCORD=true go run ./cmd/betrayal-bot/main.go

# Connect to database
sql: 
	psql $(shell cat .env | grep DATABASE_POOLER_URL | cut -d '=' -f2)

# Database migrations
migrate-up:
	migrate -database $(shell cat .env | grep DATABASE_POOLER_URL | cut -d '=' -f2) -path internal/db/migration up

migrate-down:
	migrate -database $(shell cat .env | grep DATABASE_POOLER_URL | cut -d '=' -f2) -path internal/db/migration down

migrate-sync:
	migrate -database $(shell cat .env | grep DATABASE_POOLER_URL | cut -d '=' -f2) -path internal/db/migration down && migrate -database $(shell cat .env | grep DATABASE_POOLER_URL | cut -d '=' -f2) -path internal/db/migration up

mock-migrate-up:
	migrate -database $(shell cat .env | grep MOCK_DATABASE | cut -d '=' -f2) -path internal/db/migration up

mock-migrate-down:
	migrate -database $(shell cat .env | grep MOCK_DATABASE | cut -d '=' -f2) -path internal/db/migration down

# Templ template generation
templ-generate:
	templ generate

templ-watch:
	templ generate --watch

# Tailwind CSS (requires tailwindcss standalone CLI)
tailwind-build:
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

tailwind-watch:
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch

# Generate all (templ + tailwind)
generate: templ-generate tailwind-build

# Build the binary (generates templates and CSS first)
build: generate
	go build -o ./bin/betrayal-bot ./cmd/betrayal-bot/
