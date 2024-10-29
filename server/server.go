package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiTimeout = 200 * time.Millisecond
	dbTimeout  = 10 * time.Millisecond
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func fetchCotacao(ctx context.Context) (*Cotacao, error) {

	fmt.Println("Rodando fetchCotacao")

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
		return nil, err
	}

	var data map[string]Cotacao
	err = json.Unmarshal(res, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
		return nil, err
	}

	// Obtém e imprime o valor de bid
	if cotacao, ok := data["USDBRL"]; ok {
		fmt.Println("Bid:", cotacao.Bid)
		return &cotacao, nil
	} else {
		fmt.Println("Dados da cotação não encontrados.")
		return nil, fmt.Errorf("dados da cotação não encontrados")
	}
}

func saveCotacao(ctx context.Context, db *sql.DB, cotacao *Cotacao) error {
	fmt.Println("Rodando saveCotacao")
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	query := "INSERT INTO cotacao (bid) VALUES (?)"
	_, err := db.ExecContext(ctx, query, cotacao.Bid)
	return err
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Rodando handler")

	ctx, cancel := context.WithTimeout(r.Context(), apiTimeout)
	defer cancel()

	cotacao, err := fetchCotacao(ctx)
	if err != nil {
		http.Error(w, "Erro ao obter cotação", http.StatusInternalServerError)
		log.Println("Erro ao obter cotação:", err)
		return
	}

	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		http.Error(w, "Erro ao abrir banco de dados", http.StatusInternalServerError)
		log.Println("Erro ao abrir banco de dados:", err)
		return
	}
	defer db.Close()

	ctxDB, cancelDB := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelDB()

	if err := saveCotacao(ctxDB, db, cotacao); err != nil {
		http.Error(w, "Erro ao salvar cotação no banco de dados", http.StatusInternalServerError)
		log.Println("Erro ao salvar cotação no banco de dados:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacao)
}

func createTable() error {

	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		return err
	}
	defer db.Close()

	create := `
	CREATE TABLE IF NOT EXISTS cotacao (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT
	);
	`
	_, err = db.Exec(create)
	return err
}

func main() {

	if err := createTable(); err != nil {
		log.Fatalf("Erro ao configurar banco de dados: %v", err)
	}

	http.HandleFunc("/cotacao", handler)
	fmt.Println("Servidor iniciado na porta :8080")
	http.ListenAndServe(":8080", nil)
}
