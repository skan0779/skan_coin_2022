package blockchain

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/skan0779/skan_coin_2022/utilities"
	"github.com/skan0779/skan_coin_2022/wallet"
)

type Tx struct {
	Id        string   `json:"id"`
	Timestamp int      `json:"timestamp"`
	Input     []*TxIn  `json:"input"`
	Output    []*TxOut `json:"output"`
}
type TxIn struct {
	Id        string `json:"id"`
	Index     int    `json:"index"`
	Signature string `json:"signature"`
}
type TxOut struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}
type UTxOut struct {
	Id     string `json:"id"`
	Index  int    `json:"index"`
	Amount int    `json:"amount"`
}
type mempool struct {
	Txs map[string]*Tx
	m   sync.Mutex
}

const (
	reward int = 10
)

var m *mempool
var once2 sync.Once
var ErrNotEnough = errors.New("not enough coin")
var ErrNotValid = errors.New("not valid transaction")

func (t *Tx) getId() {
	t.Id = utilities.Hash(t)
}

func (t *Tx) sign() {
	for _, txIn := range t.Input {
		txIn.Signature = wallet.Sign(wallet.Wallet(), t.Id)
	}
}

// adding transaction to mempool
func (m *mempool) AddTx(to string, amount int) (*Tx, error) {
	tx, err := makeTx(wallet.Wallet().Address, to, amount)
	if err != nil {
		return nil, err
	}
	m.Txs[tx.Id] = tx
	return tx, nil
}

// confirming transactions to Block and clear(empty) the mempool
func (m *mempool) ConfirmTx() []*Tx {
	var txs []*Tx
	coinbase := getTx(wallet.Wallet().Address)
	for _, tx := range m.Txs {
		txs = append(txs, tx)
	}
	txs = append(txs, coinbase)
	m.Txs = make(map[string]*Tx)
	return txs
}

// making coinbase transaction (blockchain -> miner)
func getTx(address string) *Tx {
	txIns := []*TxIn{
		{"", -1, "COINBASE"},
	}
	txOuts := []*TxOut{
		{address, reward},
	}
	transaction := Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		Input:     txIns,
		Output:    txOuts,
	}
	transaction.getId()
	return &transaction
}

// making transaction
func makeTx(from string, to string, amount int) (*Tx, error) {
	if CalTxOuts(from, Blockchain()) < amount {
		return nil, ErrNotEnough
	}
	var txOuts []*TxOut
	var txIns []*TxIn
	var total int = 0
	var uTxOuts []*UTxOut = FilterTxOuts(from, Blockchain())
	// create new transaction's input
	for _, uTxOut := range uTxOuts {
		if total >= amount {
			break
		}
		txIn := &TxIn{uTxOut.Id, uTxOut.Index, from}
		txIns = append(txIns, txIn)
		total += uTxOut.Amount
	}
	// create new transaction's output
	change := total - amount
	if change != 0 {
		changeTxOut := &TxOut{from, change}
		txOuts = append(txOuts, changeTxOut)
	}
	txOut := &TxOut{to, amount}
	txOuts = append(txOuts, txOut)
	// create new transaction
	tx := &Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		Input:     txIns,
		Output:    txOuts,
	}
	tx.getId()
	tx.sign()
	check := verify(tx)
	if !check {
		return nil, ErrNotValid
	}
	return tx, nil
}

// checking unspent transaction output(UTxOut) is not on mempool
func checkMempool(uTxOut *UTxOut) bool {
	for _, tx := range Mempool().Txs {
		for _, input := range tx.Input {
			if input.Id == uTxOut.Id && input.Index == uTxOut.Index {
				return true
			}
		}
	}
	return false
}

func verify(tx *Tx) bool {
	var check bool = true
	for _, txIn := range tx.Input {
		transaction := FindTx(Blockchain(), txIn.Id)
		if transaction == nil {
			check = false
			break
		}
		address := transaction.Output[txIn.Index].Address
		check = wallet.Verify(txIn.Signature, tx.Id, address)
		if !check {
			break
		}
	}
	return check
}

func GetMempool(m *mempool, rw http.ResponseWriter) {
	defer m.m.Unlock()
	m.m.Lock()
	err := json.NewEncoder(rw).Encode(m)
	utilities.ErrHandling(err)
}

func (m *mempool) UpdateTx(tx *Tx) {
	defer m.m.Unlock()
	m.m.Lock()
	m.Txs[tx.Id] = tx
}

func Mempool() *mempool {
	once2.Do(func() {
		m = &mempool{
			Txs: make(map[string]*Tx),
		}
	})
	return m
}
