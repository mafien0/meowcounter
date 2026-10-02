build:
	go build cmd/meowcounter/main.go

run:
	go run cmd/meowcounter/main.go

db-up:
	goose -dir migrations sqlite3 ./meowcounter.db up
