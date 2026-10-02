build:
	go build cmd/mecounter/main.go

run:
	go run cmd/mecounter/main.go

db-up:
	goose -dir migrations sqlite3 ./mecounter.db up
