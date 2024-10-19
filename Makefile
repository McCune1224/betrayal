run:
	go run ./cmd/betrayal-bot/main.go


.SILENT:
.PHONY: sql
sql: 
	psql $(shell cat .env | grep DATABASE_POOLER_URL | cut -d '=' -f2)

stage: 
	templ generate&& tailwindcss -i ./static/input.css -o ./static/output.css

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
