package main

import (
	"errors"
	"image/png"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"meowcounter/internal/db"
	im "meowcounter/internal/image"

	"github.com/joho/godotenv"
)

var predicate = regexp.MustCompile("^[A-Za-z0-9._,-]+$")

func main() {
	godotenv.Load()

	// -- DB --
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}
	path := os.Getenv("SQLITE_PATH")
	if path == "" {
		path = "meowcounter.db"
	}

	d, err := db.Open(db.Config{
		Driver: driver,
		Path:   path,
		URL:    os.Getenv("TURSO_DATABASE_URL"),
		Token:  os.Getenv("TURSO_AUTH_TOKEN"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	// -- API --
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("WARNING: no `PORT` env variable, defaulting to 3000")
		port = "3000"
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /@/{name}", func(w http.ResponseWriter, r *http.Request) {
		var err error
		name := r.PathValue("name")

		digitsQuery := r.URL.Query().Get("digits")
		if digitsQuery == "" {
			digitsQuery = "8"
		}
		digits, err := strconv.Atoi(digitsQuery)
		if err != nil {
			http.Error(w, "Invalid digits query!", http.StatusBadRequest)
			log.Println("400: Invalid Query")
			return
		}

		log.Printf("GET: /@/{%v}\n", name)
		if name == "" || !predicate.MatchString(name) {
			http.Error(w, "Invalid name!", http.StatusBadRequest)
			log.Println("400: Invalid Name")
			return
		}
		counter, err := d.GetCounter(r.Context(), name)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			http.Error(w, "Internal Server Error.", http.StatusInternalServerError)
			log.Println("500: %w", err)
			return
		}
		count := counter.Count + 1
		err = d.UpsertCounter(r.Context(), name, count)
		if err != nil {
			http.Error(w, "Internal Server Error.", http.StatusInternalServerError)
			log.Println("500: %w", err)
			return
		}

		img, err := im.Glue(strconv.Itoa(int(count)), digits)
		if err != nil {
			http.Error(w, "Internal Server Error.", http.StatusInternalServerError)
			log.Println("500: %w", err)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		if err := png.Encode(w, img); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("500: %w", err)
		}
	})

	log.Fatal(http.ListenAndServe(":"+port, mux))
}
