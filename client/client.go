package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	// Definição de timeout do contexto para 300 ms
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Consulta da cotação fornecida pelo servidor
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Resposta do servidor: %v", resp.Status)
	}

	var cotacao Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		log.Fatalf("Erro ao decodificar resposta: %v", err)
	}

	// Cria o arquivo
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatalf("Erro ao criar arquivo: %v", err)
	}
	defer file.Close()

	output := fmt.Sprintf("Dólar: %s", cotacao.Bid)

	// Escreve no arquivo
	_, err = file.WriteString(output)
	if err != nil {
		log.Fatalf("Erro ao escrever arquivo: %v", err)
	}

	log.Println("Cotação salva com sucesso em", "cotacao.txt")
}
