DB_URL = postgres://postgres:postgres@localhost:5432/db?sslmode=disable
MIGRATIONS_PATH = ./migrations

.PHONY: run ping test migrate-up migrate-down migrate-create

run:
	go run cmd/shortener/main.go

run-file:
	go run cmd/shortener/main.go -f="file_storage.json"

run-postgres:
	go run cmd/shortener/main.go -d="postgres://postgres:postgres@localhost:5432/db?sslmode=disable"

ping:
	curl http://localhost:8080/ping -i

test:
	go test ./...

migrate-create:
	@test -n "$(name)" || (echo "Error: name is required. Use: make migrate-create name=my_migration" && exit 1)
	migrate create -ext sql -dir "$(MIGRATIONS_PATH)" -seq "$(name)"

migrate-up:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" up

migrate-down:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" down

migrate-v:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) version

migrate-status:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" status

migrate-force:
	@test -n "$(version)" || (echo "Error: version is required. Use: make migrate-force version=1" && exit 1)
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) force $(version)

migrate-drop:
	@echo "This will DROP EVERYTHING! Continue? [y/N]" && read ans && [ $${ans:-N} = y ]
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) drop -f

migrate-reset: migrate-drop migrate-up
	@echo "Database reset and migrations reapplied"

docker-up:
	docker compose up -d --force-recreate

docker-down:
	docker compose down

docker-exec:
	docker compose exec -t postgres bash

docker-volume-rm:
	docker volume rm shorten_url_data || true 

mockgen:
	mockgen -source=internal/repository/repository.go -destination=internal/repository/mocks/mock_repository.go -package=mocks 
