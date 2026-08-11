.SILENT:
.PHONY: run run-web sql migrate-up migrate-down migrate-sync migrate-local-up migrate-local-down migrate-production-up migrate-production-down migrate-production-sync mock-migrate-up mock-migrate-down test-migration-targets templ-generate templ-watch tailwind-build tailwind-watch frontend-build build generate env-link worktree db-up db-down clean

# Extract a value from .env (handles quotes and '=' inside values, e.g. sslmode=disable)
env-value = $(shell grep -E '^$(1)=' .env | head -n1 | cut -d '=' -f2- | tr -d '"' | tr -d "'")

# Run the bot
run:
	go run ./cmd/betrayal-bot/main.go

# Run web server only (no Discord bot)
run-web:
	ENVIRONMENT=local DISABLE_DISCORD=true go run ./cmd/betrayal-bot/main.go

# Connect to the local database
sql:
	psql $(call env-value,DATABASE_URL)

# Local database migrations
migrate-local-up:
	migrate -database $(call env-value,DATABASE_URL) -path internal/db/migrate/migrations up

migrate-local-down:
	migrate -database $(call env-value,DATABASE_URL) -path internal/db/migrate/migrations down

# Compatibility aliases: preserve the historical production target names, but
# delegate to recipes that require explicit confirmation.
migrate-up: migrate-production-up

migrate-down: migrate-production-down

migrate-sync: migrate-production-sync

# Validate migration target compatibility and production refusal without running migrations.
test-migration-targets:
	./scripts/test-migration-targets.sh

# Production migrations are intentionally named and require an explicit opt-in:
#   make migrate-production-up CONFIRM_PRODUCTION_MIGRATION=YES
migrate-production-up:
	@test "$(CONFIRM_PRODUCTION_MIGRATION)" = "YES" || (echo "refusing production migration; set CONFIRM_PRODUCTION_MIGRATION=YES" >&2; exit 1)
	migrate -database $(call env-value,DATABASE_POOLER_URL) -path internal/db/migrate/migrations up

migrate-production-down:
	@test "$(CONFIRM_PRODUCTION_MIGRATION)" = "YES" || (echo "refusing production migration; set CONFIRM_PRODUCTION_MIGRATION=YES" >&2; exit 1)
	migrate -database $(call env-value,DATABASE_POOLER_URL) -path internal/db/migrate/migrations down

migrate-production-sync:
	@test "$(CONFIRM_PRODUCTION_MIGRATION)" = "YES" || (echo "refusing production migration; set CONFIRM_PRODUCTION_MIGRATION=YES" >&2; exit 1)
	migrate -database $(call env-value,DATABASE_POOLER_URL) -path internal/db/migrate/migrations down && migrate -database $(call env-value,DATABASE_POOLER_URL) -path internal/db/migrate/migrations up

mock-migrate-up:
	migrate -database $(call env-value,MOCK_DATABASE) -path internal/db/migrate/migrations up

mock-migrate-down:
	migrate -database $(call env-value,MOCK_DATABASE) -path internal/db/migrate/migrations down

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

# Generate the static SvelteKit application that Go embeds and serves.
frontend-build:
	npm --prefix frontend ci
	npm --prefix frontend run build
	node -e "const fs=require('fs');const p='internal/web/ui/dist/200.html';fs.writeFileSync(p,fs.readFileSync(p,'utf8').replace(/[\	 ]+$$/gm,''))"

# Generate all (templ + tailwind + SvelteKit static output)
generate: templ-generate tailwind-build frontend-build

# Build the binary (generates templates and CSS first)
build: generate
	go build -o ./bin/betrayal-bot ./cmd/betrayal-bot/

# Local dev database (docker compose)
db-up:
	docker compose up -d --wait

db-down:
	docker compose down

# Worktree & env tooling (see scripts/dev-env.sh)
env-link:
	./scripts/dev-env.sh link

worktree:
	./scripts/dev-env.sh new-worktree $(name)

# Remove build artifacts (never commit binaries)
clean:
	rm -rf ./bin betrayal-bot
