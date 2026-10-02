package main

import (
	"bytes"
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
		log.Printf("GET: /@/{%v}\n", name)

		// Name regex predicate
		if name == "" || !predicate.MatchString(name) {
			http.Error(w, "Invalid name!", http.StatusBadRequest)
			log.Println("400: Invalid Name")
			return
		}

		// Amount of girls holding digits
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

		// Get the counter
		counter, err := d.GetCounter(r.Context(), name)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			http.Error(w, "Internal Server Error.", http.StatusInternalServerError)
			log.Printf("500: %v\n", err)
			return
		}
		count := counter.Count + 1

		// Glue an image
		img, err := im.Glue(strconv.Itoa(int(count)), digits)
		if err != nil {
			http.Error(w, "Internal Server Error.", http.StatusInternalServerError)
			log.Printf("500: %v", err)
			return
		}

		// Store it in a buffer, so encode errors can be returned
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("500: %v", err)
		}

		// Response
		w.Header().Set("Content-Type", "image/png")
		if _, err := w.Write(buf.Bytes()); err != nil {
			log.Printf("500: %v", err)
			return
		}

		// Flush response to not keep client waiting for db write
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// slow: Update the counter
		if err := d.UpsertCounter(r.Context(), name, count); err != nil {
			log.Printf("500: %v", err)
		}
	})

	log.Fatal(http.ListenAndServe(":"+port, mux))
}
