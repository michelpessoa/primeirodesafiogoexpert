package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Bid struct {
	Bid string `json:"bid"`
}

func main() {
	contextCliente, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(contextCliente, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var bid Bid
	err = json.Unmarshal(response, &bid)
	if err != nil {
		panic(err)
	}
	fmt.Println(bid)

	f, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	_, err = f.Write([]byte("Dólar: " + bid.Bid))
	if err != nil {
		fmt.Println(err)
	}

}
