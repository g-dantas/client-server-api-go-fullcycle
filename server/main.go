package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cambio struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PCTChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	}
}

func handleRequest() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", cotacao)
	http.ListenAndServe(":8080", mux)
}

func cotacao(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cotacao, err := buscaCotacao()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao.USDBRL.Bid)
}

func buscaCotacao() (*Cambio, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var cambio Cambio
	err = json.Unmarshal(body, &cambio)
	if err != nil {
		return nil, err
	}

	insertBD(cambio.USDBRL.Bid)
	return &cambio, nil
}

func openDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		panic(err)
	}

	createTable(db)
	return db
}

func createTable(db *sql.DB) {
	sqlTableCambio := `
	CREATE TABLE IF NOT EXISTS cambio (
		id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		bid text
	);
	`
	_, err := db.Exec(sqlTableCambio)
	if err != nil {
		panic(err)
	}
}

func createDatabase() {
	_, err := os.Stat("./cotacao.db")
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll("./", 0755)
			if err != nil {
				panic(err)
			}

			database, err := os.Create("./cotacao.db")
			if err != nil {
				panic(err)
			}
			defer database.Close()
		}
	}
}

func insertBD(bid string) {
	db := openDatabase()
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.Prepare("INSERT INTO cambio(bid) values(?)")
	if err != nil {
		panic(err)
	}

	_, err = stmt.ExecContext(ctx, bid)
	if err != nil {
		panic(err)
	}
}

func main() {
	createDatabase()
	handleRequest()
}
