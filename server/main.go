package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(cotacao.USDBRL.Bid)
}

func buscaCotacao() (*Cambio, error) {
	req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	resp, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var cambio Cambio
	err = json.Unmarshal(resp, &cambio)
	if err != nil {
		return nil, err
	}

	return &cambio, nil
}

func main() {
	handleRequest()
}
