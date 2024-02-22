package transacoes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Extrato struct {
	Saldo             int
	Limite            int
	DataExtrato       time.Time          `json:"data_extrato"`
	UltimasTransacoes []TransacaoExtrato `json:"ultimas_transacoes"`
}

type Cliente struct {
	Limite int `json:"limite"`
	Saldo  int `json:"saldo"`
}

type TransacaoHandler struct {
	store *sql.DB
}

func NewTransacaoHandler(db *sql.DB) *TransacaoHandler {
	return &TransacaoHandler{store: db}
}

type Transacao struct {
	TransacaoId int    `json:"transacao_id"`
	Valor       int    `json:"valor"`
	Tipo        string `json:"tipo"`
	Descricao   string `json:"descricao"`
	ClienteId   int    `json:"cliente_id"`
}

type TransacaoExtrato struct {
	Valor       int       `json:"valor"`
	Tipo        string    `json:"tipo"`
	Descricao   string    `json:"descricao"`
	RealizadaEm time.Time `json:"realizada_em"`
}

const getClienteByIdSQL = "SELECT limite, saldo from clientes WHERE cliente_id = $1"

func (t *TransacaoHandler) findClientById(id int) (Cliente, error) {
	cliente := Cliente{}
	if err := t.store.QueryRow(getClienteByIdSQL, id).Scan(&cliente.Limite, &cliente.Saldo); err != nil {
		return cliente, err
	}
	return cliente, nil
}

const updateClienteSaldoSQL = "UPDATE clientes SET saldo = $1 WHERE cliente_id = $2"

func (t *TransacaoHandler) updateSaldoCliente(novoSaldo int, id int) bool {
	if _, err := t.store.Exec(updateClienteSaldoSQL, novoSaldo, id); err != nil {
		fmt.Printf("%v\n", err)
		return false
	}
	return true
}

const createTransacaoSQL = "INSERT INTO transacoes (valor, tipo, descricao, cliente_id) VALUES ($1, $2, $3, $4)"

func (t *TransacaoHandler) createTransacao(transacao Transacao) bool {
	if _, err := t.store.Exec(
		createTransacaoSQL,
		transacao.Valor,
		transacao.Tipo,
		transacao.Descricao,
		transacao.ClienteId); err != nil {
		return false
	}
	return true
}

const getExtratoFromClienteByIdSQL = `
SELECT 
    c.limite AS limite,
    c.saldo AS saldo,
    t.valor,
    t.tipo,
    t.descricao,
    t.realizada_em
FROM clientes c
JOIN transacoes t ON c.cliente_id = t.cliente_id
WHERE c.cliente_id = $1
ORDER BY t.realizada_em DESC`

func (t *TransacaoHandler) getExtratoFromClienteById(id int) (Extrato, error) {
	extrato := Extrato{DataExtrato: time.Now()}
	if err := t.store.QueryRow("SELECT limite, saldo from clientes where cliente_id = $1", id).Scan(&extrato.Limite, &extrato.Saldo); err != nil {
		return extrato, err
	}

	rows, err := t.store.Query("SELECT valor, tipo, descricao, realizada_em from transacoes where cliente_id = $1", id)

	if err != nil {
		fmt.Println("1")
		return extrato, err
	}

	defer rows.Close()

	for rows.Next() {
		var transacao TransacaoExtrato
		if err := rows.Scan(&transacao.Valor, &transacao.Tipo, &transacao.Descricao, &transacao.RealizadaEm); err != nil {
			fmt.Println("2")

			return extrato, err
		}
		extrato.UltimasTransacoes = append(extrato.UltimasTransacoes, transacao)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("3")
		return extrato, err
	}

	return extrato, nil
}

func (t *TransacaoHandler) validateTransaction(transacao *Transacao) bool {
	if len(transacao.Descricao) < 1 || len(transacao.Descricao) > 10 {
		return false
	}
	if transacao.Tipo != "d" && transacao.Tipo != "c" {
		return false
	}
	return true
}

func unnaprovedTransaction(w http.ResponseWriter) {
	w.WriteHeader(422)
	w.Write([]byte("Transacao reprovada"))
}

func validationError(w http.ResponseWriter) {
	w.WriteHeader(422)
	w.Write([]byte("Erro de validacao"))
}

func badId(w http.ResponseWriter) {
	w.WriteHeader(404)
	w.Write([]byte("Sem id"))
}

func dataNotFound(w http.ResponseWriter) {
	w.WriteHeader(404)
	w.Write([]byte("Dado nao encontrado"))
}

func internalError(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("Internal Error"))
}

func (t *TransacaoHandler) extrato(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.Split(r.URL.Path, "/")[2])

	if err != nil {
		badId(w)
		return
	}

	extrato, err := t.getExtratoFromClienteById(id)

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Erro na transacao de extrato"))
		return
	}

	extratoResponse, err := json.Marshal(&extrato)

	if err != nil {
		internalError(w)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(extratoResponse)
}

func (t *TransacaoHandler) transacao(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.Split(r.URL.Path, "/")[2])

	if err != nil {
		badId(w)
		return
	}

	transacao := Transacao{ClienteId: id}

	err = json.NewDecoder(r.Body).Decode(&transacao)

	if err != nil {
		validationError(w)
		return
	}

	defer r.Body.Close()

	if ok := t.validateTransaction(&transacao); !ok {
		validationError(w)
		return
	}

	cliente, err := t.findClientById(id)

	fmt.Println("Transacao: findClientById ", cliente)

	if err != nil {
		dataNotFound(w)
		return
	}

	var newSaldo int

	if transacao.Tipo == "d" {
		newSaldo = cliente.Saldo - transacao.Valor
		if newSaldo < (cliente.Limite * -1) {
			unnaprovedTransaction(w)
			return
		}

		cliente.Saldo = newSaldo

		if ok := t.updateSaldoCliente(newSaldo, id); !ok {
			internalError(w)
			return
		}

	} else if transacao.Tipo == "c" {
		fmt.Println("not implemented yet!")
	}

	if ok := t.createTransacao(transacao); !ok {
		internalError(w)
		return
	}

	clienteResponseData, err := json.Marshal(cliente)

	if err != nil {
		internalError(w)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(clienteResponseData)
}

func (t *TransacaoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		t.transacao(w, r)
	case "GET":
		t.extrato(w, r)
	}
}
