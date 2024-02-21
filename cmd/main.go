package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/lucasdellatorre/rinha-de-backend-2024-q1/internal/transacoes"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	db, err := setupDatabase()

	if err != nil {
		db.Close()
		return err
	}

	fmt.Println("DB Connected!")

	defer db.Close()

	router := http.NewServeMux()

	router.Handle("/clientes/", transacoes.NewTransacaoHandler(db))

	s := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Server is running...")
	return s.ListenAndServe()
}

func setupDatabase() (*sql.DB, error) {
	connStr := "postgres://admin:123@db:5432/rinha?sslmode=disable"

	var db *sql.DB
	var err error

	// Retry connecting to the database with exponential backoff
	retries := 0
	maxRetries := 5
	for db == nil {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			retries++
			if retries >= maxRetries {
				return nil, errors.New("failed to connect to the database after maximum retries")
			}
			fmt.Printf("Failed to connect to the database (attempt %d/%d). Retrying in 5 seconds...\n", retries, maxRetries)
			time.Sleep(5 * time.Second)
		}
	}

	// Check if the database is ready by pinging it
	for {
		if err := db.Ping(); err != nil {
			fmt.Println("Error pinging the database. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	return db, nil
}
