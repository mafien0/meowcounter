package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"mecounter/internal/db"

	"github.com/joho/godotenv"
)

var predicate = regexp.MustCompile("^[A-Za-z0-9._,-]+$")

func main() {
	godotenv.Load()

	// -- DB --
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Println("WARNING: no `DB_PATH` env variable, defaulting to `<cwd>/mecounter.db`")
		dbPath = "mecounter.db"
	}

	d, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	// -- API --
	port := os.Getenv("PORT")
	if dbPath == "" {
		log.Println("WARNING: no `PORT` env variable, defaulting to 3000")
		dbPath = "3000"
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /@/{name}", func(w http.ResponseWriter, r *http.Request) {
		var err error
		name := r.PathValue("name")
		log.Printf("GET: /@/{%v}\n", name)
		if name == "" || !predicate.MatchString(name) {
			http.Error(w, "Invalid name!", http.StatusBadRequest)
			return
		}
		counter, err := d.GetCounter(r.Context(), name)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			http.Error(w, "Internal Server Error.", http.StatusInternalServerError)
			return
		}
		count := counter.Count + 1
		err = d.UpsertCounter(r.Context(), name, count)
		if err != nil {
			http.Error(w, "Internal Server Error.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(strconv.FormatInt(count, 10)))
	})

	log.Fatal(http.ListenAndServe(":"+port, mux))
}
