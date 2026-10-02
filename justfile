set dotenv-load := true

db_driver := env("DB_DRIVER", "sqlite")
sqlite_path := env("SQLITE_PATH", "meowcounter.db")

build:
	go build cmd/meowcounter/main.go

run:
	go run cmd/meowcounter/main.go

db-up:
	@if [ "{{db_driver}}" = "turso" ]; then \
		goose -dir migrations turso "$TURSO_DATABASE_URL?authToken=$TURSO_AUTH_TOKEN" up; \
	else \
		goose -dir migrations sqlite3 "{{sqlite_path}}" up; \
	fi

db-status:
	@if [ "{{db_driver}}" = "turso" ]; then \
		goose -dir migrations turso "$TURSO_DATABASE_URL?authToken=$TURSO_AUTH_TOKEN" status; \
	else \
		goose -dir migrations sqlite3 "{{sqlite_path}}" status; \
	fi
