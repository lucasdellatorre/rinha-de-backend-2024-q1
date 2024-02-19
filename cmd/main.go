package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Cliente struct {
	Id        int    `json:"id"`
	Valor     int    `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

type clientesHandler struct{}

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

	fmt.Println(cliente)
}

func main() {
	mux := http.NewServeMux()

	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	mux.Handle("/clientes/", &clientesHandler{})

	fmt.Println("Server is running...")
	log.Fatal(s.ListenAndServe())
}
