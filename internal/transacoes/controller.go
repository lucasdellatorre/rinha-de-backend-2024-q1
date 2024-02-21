package transacoes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const getClienteById = "SELECT limite, saldo from clientes WHERE cliente_id = $1"

const updateClienteSaldoLimite = "UPDATE clientes SET saldo = $1, limite = $2 WHERE cliente_id = $3"

func (t *TransacaoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientId, err := strconv.Atoi(strings.Split(r.URL.Path, "/")[2])

	if err != nil {
		fmt.Println("Invalid id:", err)
		return
	}

	var limite int
	var saldo int

	if err := t.store.QueryRow(getClienteById, clientId).Scan(&limite, &saldo); err != nil {
		fmt.Println(err)
		fmt.Println("Erro ao pegar limite e saldo do cliente")
	}

	// nota: estudar sobre transacoes, pq nao estou fazendo a validacao pelo banco de dados e em multiplas requisicoes concorrentes
	// ira ter problemas de concistencia. Qual nivel de consistencia escolher?

	fmt.Println("limite", limite)
	fmt.Println("saldo", saldo)

	transacao := Transacao{ClienteId: clientId}

	json.NewDecoder(r.Body).Decode(&transacao)

	defer r.Body.Close()

	if transacao.Tipo == "d" {
		newSaldo := saldo - transacao.Valor
		newLimite := limite - transacao.Valor
		if newSaldo < newLimite {
			fmt.Println("Erro 422")
		}

		if err := t.store.QueryRow(updateClienteSaldoLimite, newSaldo, newLimite, clientId); err != nil {
			fmt.Println(err)
		}
	}

	if err := t.store.QueryRow("INSERT INTO transacoes (valor, tipo, descricao, cliente_id) VALUES ($1, $2, $3, $4)",
		transacao.Valor, transacao.Tipo, transacao.Descricao, transacao.ClienteId); err != nil {
		fmt.Println(err)
	}
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
