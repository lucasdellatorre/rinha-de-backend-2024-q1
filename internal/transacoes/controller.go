package transacoes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Cliente struct {
	limite int
	saldo  int
}

const getClienteByIdSQL = "SELECT limite, saldo from clientes WHERE cliente_id = $1"

func (t *TransacaoHandler) findClientById(ctx context.Context, id int) (Cliente, error) {
	cliente := Cliente{}
	if err := t.store.QueryRow(getClienteByIdSQL, id).Scan(&cliente.limite, &cliente.saldo); err != nil {
		return cliente, err
	}
	return cliente, nil
}

const updateClienteSaldoSQL = "UPDATE clientes SET saldo = $1 WHERE cliente_id = $2"

func (t *TransacaoHandler) updateSaldoCliente(ctx context.Context, novoSaldo int, id int) bool {
	if err := t.store.QueryRow(updateClienteSaldoSQL, novoSaldo, id); err != nil {
		return false
	}
	return true
}

const createTransacaoSQL = "INSERT INTO transacoes (valor, tipo, descricao, cliente_id) VALUES ($1, $2, $3, $4)"

func (t *TransacaoHandler) createTransacao(ctx context.Context, transacao Transacao) bool {
	if err := t.store.QueryRow(
		createTransacaoSQL,
		transacao.Valor,
		transacao.Tipo,
		transacao.Descricao,
		transacao.ClienteId); err != nil {
		return false
	}
	return true
}

func (t *TransacaoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	// tx, err := t.store.BeginTx(ctx, nil)

	clientId, err := strconv.Atoi(strings.Split(r.URL.Path, "/")[2])

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Sem id"))
		return
	}

	cliente, err := t.findClientById(ctx, clientId)

	// nota: estudar sobre transacoes, pq nao estou fazendo a validacao pelo banco de dados e em multiplas requisicoes concorrentes
	// ira ter problemas de concistencia. Qual nivel de consistencia escolher?

	transacao := Transacao{ClienteId: clientId}

	json.NewDecoder(r.Body).Decode(&transacao)

	defer r.Body.Close()

	var newSaldo int

	if transacao.Tipo == "d" {
		newSaldo = cliente.saldo - transacao.Valor
		if newSaldo < (cliente.limite * -1) {
			w.WriteHeader(422)
			w.Write([]byte("Erro 422"))
			return
		}

		cliente.saldo = newSaldo

		if ok := t.updateSaldoCliente(ctx, newSaldo, clientId); !ok {
			fmt.Println("Erro ao atualizar novo saldo do cliente")
			return
		}

	} else if transacao.Tipo == "c" {

	}

	if ok := t.createTransacao(ctx, transacao); !ok {
		fmt.Println("Erro ao criar transacao")
		return
	}

	clienteResponseData, err := json.Marshal(cliente)

	if err != nil {
		fmt.Println("Erro json marshalling")
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(clienteResponseData)
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
