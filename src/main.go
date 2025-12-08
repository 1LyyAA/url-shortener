//go:build !solution

package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Request struct {
	Url string `json:"url"`
}

type Response struct {
	Url      string `json:"url"`
	Key      string `json:"key"`
	ShortUrl string `json:"short_url"`
}

var db *sql.DB

func initDB() error {
	var err error
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	connStr := fmt.Sprintf("host=%s port=5432 user=admin password=admin dbname=db sslmode=disable", host)
	db, err = sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	// Create table if not exists
	schema := `
	CREATE TABLE IF NOT EXISTS urls (
		key VARCHAR(8) PRIMARY KEY,
		url TEXT NOT NULL UNIQUE
	);
	CREATE INDEX IF NOT EXISTS idx_url ON urls(url);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("unable to create table: %w", err)
	}

	return nil
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		requestPath := r.URL.Path
		key := path.Base(requestPath)

		var url string
		err := db.QueryRow("SELECT url FROM urls WHERE key = $1", key).Scan(&url)
		if err == sql.ErrNoRows {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		// Ensure URL has a scheme
		if len(url) > 0 && url[0] != 'h' {
			url = "http://" + url
		}

		http.Redirect(w, r, url, http.StatusMovedPermanently)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		req, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var request Request
		if err := json.Unmarshal(req, &request); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Check if URL already exists
		var existingKey string
		err = db.QueryRow("SELECT key FROM urls WHERE url = $1", request.Url).Scan(&existingKey)
		if err == nil {
			// URL already exists, return existing key
			host := r.Host
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			shortUrl := fmt.Sprintf("%s://%s/go/%s", scheme, host, existingKey)
			response := Response{Url: request.Url, Key: existingKey, ShortUrl: shortUrl}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if err != sql.ErrNoRows {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		// Generate new key
		var key string
		for {
			b := make([]byte, 4)
			rand.Read(b)
			key = hex.EncodeToString(b)

			// Try to insert, if key collision happens, generate new one
			_, err = db.Exec("INSERT INTO urls (key, url) VALUES ($1, $2)", key, request.Url)
			if err == nil {
				break
			}
			// Continue loop to generate new key on collision
		}

		host := r.Host
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		shortUrl := fmt.Sprintf("%s://%s/go/%s", scheme, host, key)
		response := Response{Url: request.Url, Key: key, ShortUrl: shortUrl}
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	port := flag.Int("port", 8080, "Port flag")
	flag.Parse()

	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	http.HandleFunc("/shorten", PostHandler)
	http.HandleFunc("/go/", GetHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("Starting server on %s", addr)
	http.ListenAndServe(addr, nil)
}
