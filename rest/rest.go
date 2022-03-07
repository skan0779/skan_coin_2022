// Package rest provides Rest Api for skancoin
package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skan0779/skan_coin_2022/blockchain"
	"github.com/skan0779/skan_coin_2022/p2p"
	"github.com/skan0779/skan_coin_2022/utilities"
	"github.com/skan0779/skan_coin_2022/wallet"
)

var port string

type url string
type url_data struct {
	URL         url    `json:"url"`
	Method      string `json:"method"`
	Description string `json:"description"`
}
type errResponse struct {
	Message string `json:"errorMessage"`
}
type balanceResponse struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}
type addTxResponse struct {
	To     string
	Amount int
}
type walletResponse struct {
	Address string `json:"address"`
}
type connectResponse struct {
	Address string `json:"address"`
	Port    string `json:"port"`
}

func (u url_data) String() string {
	return "URL Description"
}

func (u url) MarshalText() ([]byte, error) {
	url := fmt.Sprintf("http://localhost%s%s", port, u)
	return []byte(url), nil
}

func documentation(rw http.ResponseWriter, r *http.Request) {
	data := []url_data{
		{URL: url("/"), Method: "GET", Description: "look up document"},
		{URL: url("/blocks"), Method: "GET", Description: "load blocks"},
		{URL: url("/blocks"), Method: "POST", Description: "add a new block"},
		{URL: url("/blocks/{hash}"), Method: "GET", Description: "load a block"},
		{URL: url("/status"), Method: "GET", Description: "load a status"},
		{URL: url("/balance/{address}"), Method: "GET", Description: "load a balance"},
		{URL: url("/mempool"), Method: "GET", Description: "load a memory pool"},
		{URL: url("/transactions"), Method: "GET", Description: "load a transactions"},
		{URL: url("/transactions"), Method: "POST", Description: "add a transaction"},
		{URL: url("/wallet"), Method: "GET", Description: "load a wallet"},
		{URL: url("/ws"), Method: "GET", Description: "upgrade to websocket"},
		{URL: url("/connect"), Method: "GET", Description: "load peers"},
		{URL: url("/connect"), Method: "GET", Description: "add a peer"},
	}
	json.NewEncoder(rw).Encode(data)
}

func blocks(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		json.NewEncoder(rw).Encode(blockchain.Blocks(blockchain.Blockchain()))
	case "POST":
		block := blockchain.Blockchain().AddBlock()
		p2p.BroadcastBlock(block)
		rw.WriteHeader(http.StatusCreated)
	}
}

func block(rw http.ResponseWriter, r *http.Request) {
	// get the value from request body
	vars := mux.Vars(r)
	hash := vars["hash"]
	// find block with height (received value)
	encoder := json.NewEncoder(rw)
	block, err2 := blockchain.FindBlock(hash)
	// error handling: if the finding height is not in range
	if err2 == blockchain.ErrNotFound {
		encoder.Encode(errResponse{Message: fmt.Sprint(err2)})
	} else {
		encoder.Encode(block)
	}
}

func status(rw http.ResponseWriter, r *http.Request) {
	blockchain.GetStatus(blockchain.Blockchain(), rw)
}

func balance(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	total := r.URL.Query().Get("total")
	switch total {
	case "true":
		amount := blockchain.CalTxOuts(address, blockchain.Blockchain())
		json.NewEncoder(rw).Encode(balanceResponse{
			Address: address,
			Balance: amount,
		})
	default:
		utilities.ErrHandling(json.NewEncoder(rw).Encode(blockchain.FilterTxOuts(address, blockchain.Blockchain())))
	}
}

func mempool(rw http.ResponseWriter, r *http.Request) {
	blockchain.GetMempool(blockchain.Mempool(), rw)
}

func transactions(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		utilities.ErrHandling(json.NewEncoder(rw).Encode(blockchain.Mempool().Txs))
	case "POST":
		var response addTxResponse
		utilities.ErrHandling(json.NewDecoder(r.Body).Decode(&response))
		tx, err := blockchain.Mempool().AddTx(response.To, response.Amount)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(rw).Encode(errResponse{err.Error()})
			return
		}
		p2p.BroadcastTx(tx)
		rw.WriteHeader(http.StatusCreated)
	}
}

func myWallet(rw http.ResponseWriter, r *http.Request) {
	address := wallet.Wallet().Address
	json.NewEncoder(rw).Encode(walletResponse{Address: address})
}

func connect(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		json.NewEncoder(rw).Encode(p2p.GetPeers(&p2p.Peers))
	case "POST":
		var res *connectResponse
		json.NewDecoder(r.Body).Decode(&res)
		p2p.ConnNode(res.Address, res.Port, port[1:], true)
		rw.WriteHeader(http.StatusOK)
	}
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(rw, r)
	})
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s", r.URL.Path)
		next.ServeHTTP(rw, r)
	})
}

// start function with the port code
func Start(p int) {
	port = fmt.Sprintf(":%d", p)
	handler := mux.NewRouter()
	handler.Use(jsonMiddleware, loggerMiddleware)
	handler.HandleFunc("/", documentation).Methods("GET")
	handler.HandleFunc("/blocks", blocks).Methods("GET", "POST")
	handler.HandleFunc("/blocks/{hash:[a-f0-9]+}", block).Methods("GET")
	handler.HandleFunc("/status", status).Methods("GET")
	handler.HandleFunc("/balance/{address}", balance).Methods("GET")
	handler.HandleFunc("/mempool", mempool).Methods("GET")
	handler.HandleFunc("/transactions", transactions).Methods("GET", "POST")
	handler.HandleFunc("/wallet", myWallet).Methods("GET")
	handler.HandleFunc("/ws", p2p.Upgrade).Methods("GET")
	handler.HandleFunc("/connect", connect).Methods("GET", "POST")
	fmt.Printf("Start REST API %s \n", port)
	log.Fatal(http.ListenAndServe(port, handler))
}
