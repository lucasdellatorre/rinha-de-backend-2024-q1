package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Cliente struct {
	Id        int    `json:"id"`
	Valor     int    `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

type clientesHandler struct {
	store *sql.DB
}

func (c *clientesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientId, err := strconv.Atoi(strings.Split(r.URL.Path, "/")[2])

	if err != nil {
		fmt.Println("Invalid id:", err)
		return
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	cliente := Cliente{Id: clientId}

	json.Unmarshal([]byte(body), &cliente)

	result, err := c.store.Exec("INSERT INTO clientes (name, limite) VALUES (?, ?)", cliente.Descricao, cliente.Valor)

	if err != nil {
		fmt.Println("error while inserting")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Erro pegando o id")
		return
	}

	fmt.Println("novo id: ", id)

	fmt.Println(cliente)
}

func main() {
	connStr := "postgres://admin:123@db:5432/rinha"

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
				log.Fatal("Failed to connect to the database after maximum retries")
			}
			fmt.Printf("Failed to connect to the database (attempt %d/%d). Retrying in 5 seconds...\n", retries, maxRetries)
			time.Sleep(3 * time.Second)
		}
	}

	defer db.Close()

	// Check if the database is ready by pinging it
	for {
		if err := db.Ping(); err != nil {
			fmt.Println("Error pinging the database. Retrying in 5 seconds...")
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}

	fmt.Println("DB Connected!")

	mux := http.NewServeMux()

	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	mux.Handle("/clientes/", &clientesHandler{
		store: db,
	})

	fmt.Println("Server is running...")
	log.Fatal(s.ListenAndServe())
}
