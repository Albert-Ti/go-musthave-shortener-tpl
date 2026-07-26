DB_URL = postgres://postgres:postgres@localhost:5432/db?sslmode=disable
MIGRATIONS_PATH = ./migrations
PPROF_FILE_PATH = profiles/base.pprof

.PHONY: run ping test migrate-up migrate-down migrate-create

# Запуск сервера с настройками по умолчанию (in-memory хранилище)
run:
	go run cmd/shortener/main.go

# Запуск сервера с файловым хранилищем
run-file:
	go run cmd/shortener/main.go -f="file_storage.json"

# Запуск сервера с Postgres и файлом аудита
run-pg:
	go run cmd/shortener/main.go -d="postgres://postgres:postgres@localhost:5432/db?sslmode=disable" --audit-file="audit.json"

# То же, что run-pg, но с включённым pprof-сервером (MODE=debug)
run-pg-debug:
	export MODE=debug && go run cmd/shortener/main.go -d="postgres://postgres:postgres@localhost:5432/db?sslmode=disable" --audit-file="audit.json"

# Проверка доступности сервера
ping:
	curl http://localhost:8080/ping -i

# Запуск всех тестов
test:
	go test ./...

# Создание новой миграции: make migrate-create name=my_migration
migrate-create:
	@test -n "$(name)" || (echo "Error: name is required. Use: make migrate-create name=my_migration" && exit 1)
	migrate create -ext sql -dir "$(MIGRATIONS_PATH)" -seq "$(name)"

# Применить все миграции
migrate-up:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" up

# Откатить все миграции
migrate-down:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" down

# Показать текущую версию миграции
migrate-v:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) version

# Показать статус миграций
migrate-status:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" status

# Принудительно выставить версию миграции без её применения: make migrate-force version=1
migrate-force:
	@test -n "$(version)" || (echo "Error: version is required. Use: make migrate-force version=1" && exit 1)
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) force $(version)

# Удалить все таблицы из базы (требует подтверждения)
migrate-drop:
	@echo "This will DROP EVERYTHING! Continue? [y/N]" && read ans && [ $${ans:-N} = y ]
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) drop -f

# Полный сброс базы: drop + повторное применение миграций
migrate-reset: migrate-drop migrate-up
	@echo "Database reset and migrations reapplied"

# Поднять контейнеры (с пересозданием)
docker-up:
	docker compose up -d --force-recreate

# Остановить контейнеры
docker-down:
	docker compose down

# Зайти в контейнер postgres по bash
docker-exec:
	docker compose exec -t postgres bash

# Удалить том с данными Postgres
docker-volume-rm:
	docker volume rm shorten_url_data || true 

# Сгенерировать моки репозитория через mockgen
mockgen:
	mockgen -source=internal/repository/repository.go -destination=internal/repository/mocks/mock_repository.go -package=mocks 

# Нагрузочный тест: 1000 запросов, 10 одновременных соединений
hey:
	hey -n 1000 -c 10 http://localhost:8080/

# Снять heap-профиль pprof в файл: make pprof-snapshot file_path=profiles/x.pprof
pprof-snapshot:
	@test -n "$(file_path)" || (echo "Error: file_path is required. Use: make  file_path=profiles/file_path.pprof" && exit 1)
	curl -s -k -v http://localhost:6060/debug/pprof/heap > "$(file_path)"

# Открыть профиль в веб-интерфейсе pprof: make pprof-run-web file_path=profiles/x.pprof
pprof-run-web:
	@test -n "$(file_path)" || (echo "Error: file_path is required. Use: make  file_path=profiles/file_path.pprof" && exit 1)
	go tool pprof -http=":9090" "$(file_path)"

# Сравнить base.pprof и result.pprof (разница по памяти)
pprof-diff:
	go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof 

# Запустить локальный сервер документации pkgsite на :6000
pkgsite:
	pkgsite -http=":6000" .

# Скомпилировать multichecker.Main и запустить анализ через набор игнорируемых правил staticcheck_ignore.json
analyze:
	go build -o ./staticlint ./cmd/staticlint/ && ./staticlint ./...

# Запуск анализа через файл конфигурации staticcheck.conf
analyze-conf:
	staticcheck ./...

# Кодогенерация - просканирует все файлы текущей директории и запустит операции, указанные в комментариях //go:generate.
generate:
	go generate ./... 