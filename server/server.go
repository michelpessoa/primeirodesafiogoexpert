package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
)

type CambioConsulta struct {
	Usdbrl Usdbrl `json:"USDBRL"`
}

type Usdbrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type CambioResposta struct {
	ValorDolar string `json:"bid"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cambio", CambioHandler)

	http.ListenAndServe(":8080", mux)
}

func CambioHandler(w http.ResponseWriter, r *http.Request) {
	context, cancel := context.WithTimeout(r.Context(), time.Millisecond*210)
	defer cancel()

	//time.Sleep(1 * time.Second) //Teste para o timeout

	select {
	case <-context.Done():
		w.WriteHeader(http.StatusRequestTimeout)
	default:

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		consulta, err := BuscaCambio()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		context2, cancel := context.WithTimeout(r.Context(), time.Millisecond*10)
		defer cancel()

		time.Sleep(1 * time.Second) //Teste para o timeout

		select {
		case <-context2.Done():
			w.WriteHeader(http.StatusGatewayTimeout)

		default:
			SalvaConsultaCambio(consulta.Usdbrl)

		}

		resposta := CambioResposta{consulta.Usdbrl.Bid}

		json.NewEncoder(w).Encode(resposta)
	}

}

func BuscaCambio() (*CambioConsulta, error) {
	resp, error := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if error != nil {
		return nil, error
	}
	defer resp.Body.Close()
	body, error := io.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}

	var c CambioConsulta
	error = json.Unmarshal(body, &c)
	if error != nil {
		return nil, error
	}
	fmt.Println(c.Usdbrl.Bid)

	return &c, nil
}

func SalvaConsultaCambio(consulta Usdbrl) error {
	db, err := gorm.Open(sqlite.Open("cambio.db"), &gorm.Config{})
	//db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Usdbrl{})
	err = db.Create(consulta).Error
	if err != nil {
		return err
	}
	return nil

}
