CREATE TABLE clientes (
    cliente_id SERIAL PRIMARY KEY,
    limite INTEGER NOT NULL,
    saldo INTEGER NOT NULL
);

CREATE TABLE transacoes (
    id SERIAL PRIMARY KEY,
    valor INTEGER NOT NULL,
    tipo CHAR(1) NOT NULL,
    descricao VARCHAR(10) NOT NULL,
    realizada_em timestamptz NOT NULL DEFAULT (now()),
    FOREIGN KEY (cliente_id) REFERENCES clientes (cliente_id)
);

ALTER TABLE transacoes
INSERT INTO clientes (limite, saldo)
VALUES
    (100000, 0),
    (80000, 0),
    (1000000, 0),
    (10000000, 0),
    (500000, 0);
