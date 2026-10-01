build:
	go build cmd/mebuild/main.go

run:
	go run cmd/mebuild/main.go

db-up:
	goose -dir migrations sqlite3 ./mecounter.db up
