package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
)

type CotacaoConsulta struct {
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

type CotacaoResposta struct {
	ValorDolar string `json:"bid"`
}

type RequestError struct {
	StatusCode int

	Err error
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)

	http.ListenAndServe(":8080", mux)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	context, cancel := context.WithTimeout(r.Context(), time.Millisecond*210)
	defer cancel()

	consulta, err := BuscaCotacao(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	erroDB := SalvaConsultaCotacao(consulta.Usdbrl, r)
	if erroDB != nil {
		fmt.Println(erroDB)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//time.Sleep(220 * time.Millisecond) //Teste para o timeout

	select {
	case <-context.Done():
		fmt.Println("Fatal Error")
		w.WriteHeader(http.StatusInternalServerError)
	default:

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resposta := CotacaoResposta{consulta.Usdbrl.Bid}

		json.NewEncoder(w).Encode(resposta)
	}

}

func BuscaCotacao(r *http.Request) (*CotacaoConsulta, error) {
	context1, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
	defer cancel()
	resp, error := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")

	//time.Sleep(201 * time.Millisecond) //Teste para o timeout

	select {
	case <-context1.Done():
		return nil, &RequestError{
			StatusCode: 408,
			Err:        errors.New(" timeout request"),
		}
	default:
		if error != nil {
			return nil, error
		}
		defer resp.Body.Close()
		body, error := io.ReadAll(resp.Body)
		if error != nil {
			return nil, error
		}

		var c CotacaoConsulta
		error = json.Unmarshal(body, &c)
		if error != nil {
			return nil, error
		}
		fmt.Println(c.Usdbrl.Bid)

		return &c, nil
	}

}

func SalvaConsultaCotacao(consulta Usdbrl, r *http.Request) error {
	context2, cancel := context.WithTimeout(r.Context(), time.Millisecond*10)
	defer cancel()

	//time.Sleep(20 * time.Millisecond) //Teste para o timeout

	select {
	case <-context2.Done():
		return &RequestError{
			StatusCode: 408,
			Err:        errors.New(" timeout transaction"),
		}
	default:
		db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
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

}

func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Err)
}
