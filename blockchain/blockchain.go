// Package blockchain provides all the blockchain functions for skancoin
package blockchain

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/skan0779/skan_coin_2022/db"
	"github.com/skan0779/skan_coin_2022/utilities"
)

type blockchain struct {
	Height     int    `json:"height"`
	PreHash    string `json:"prevHash,omitempty"`
	Difficulty int    `json:"difficulty"`
	m          sync.Mutex
}

var b *blockchain
var once sync.Once
var ErrNotFound = errors.New("not found")

const (
	difficultyDefault  int = 2
	difficultyInterval int = 5
	blockInterval      int = 2
	toleranceInterval  int = 2
)

// add the new block
func (b *blockchain) AddBlock() *Block {
	block := createBlock(b.Height+1, b.PreHash, difficulty(b))
	b.Height = block.Height
	b.PreHash = block.Hash
	b.Difficulty = block.Difficulty
	saveBlockchain(b)
	return block
}

// update the blocks from other node
func (b *blockchain) Update(blocks []*Block) {
	defer b.m.Unlock()
	b.m.Lock()
	b.Difficulty = blocks[0].Difficulty
	b.Height = len(blocks)
	b.PreHash = blocks[0].Hash
	saveBlockchain(b)
	db.UpdateBucketBlocks()
	for _, block := range blocks {
		saveBlock(block)
	}
}

// update the block and mempool from the other node
func (b *blockchain) UpdateBlock(block *Block) {
	defer b.m.Unlock()
	b.m.Lock()
	b.Height += 1
	b.Difficulty = block.Difficulty
	b.PreHash = block.Hash
	saveBlockchain(b)
	saveBlock(block)

	for _, tx := range block.Transactions {
		_, ok := m.Txs[tx.Id]
		if ok {
			delete(m.Txs, tx.Id)
		}
	}
}

// get the blocks
func Blocks(b *blockchain) []*Block {
	defer b.m.Unlock()
	b.m.Lock()
	var blocks []*Block
	hash := b.PreHash
	for {
		block, _ := FindBlock(hash)
		blocks = append(blocks, block)
		if block.PreHash != "" {
			hash = block.PreHash
		} else {
			break
		}
	}
	return blocks
}

// save to db blockchain bucket
func saveBlockchain(b *blockchain) {
	db.SaveBlockchain(utilities.ToByte(b))
}

// calculate and set the difficulty
func difficulty(b *blockchain) int {
	if b.Height == 0 {
		return difficultyDefault
	} else if b.Height%difficultyInterval == 0 {
		// recalculate difficulty
		blocks := Blocks(b)
		lastBlock := blocks[0]
		lastCalBlock := blocks[difficultyInterval-1]
		actualTime := (lastBlock.Timestamp / 60) - (lastCalBlock.Timestamp / 60)
		expectTime := blockInterval * difficultyInterval
		if actualTime <= (expectTime - toleranceInterval) {
			return b.Difficulty + 1
		} else if actualTime >= (expectTime + toleranceInterval) {
			return b.Difficulty - 1
		}
		return b.Difficulty
	} else {
		return b.Difficulty
	}
}

// filter all the unspent transaction outputs by address
func FilterTxOuts(address string, b *blockchain) []*UTxOut {
	var uTxOuts []*UTxOut
	sTxOuts := make(map[string]bool)

	for _, block := range Blocks(b) {
		for _, tx := range block.Transactions {
			for _, input := range tx.Input {
				// if TxIn is a coinbase transaction: blockchain mining transaction 은 계산 x
				if input.Signature == "COINBASE" {
					break
				}
				// if TxIn is p2p transaction: 거래할때 사용한 코인들 채크표시
				// mark that transaction id is used
				if FindTx(b, input.Id).Output[input.Index].Address == address {
					sTxOuts[input.Id] = true
				}
			}
			for i, output := range tx.Output {
				if output.Address == address {
					// if the transaction outputs not been used in other transaction input
					_, check := sTxOuts[tx.Id]
					if !check {
						uTxOut := &UTxOut{tx.Id, i, output.Amount}
						// if the transaction outputs not been used in memory pool input
						if !checkMempool(uTxOut) {
							uTxOuts = append(uTxOuts, uTxOut)
						}
					}
				}
			}
		}
	}
	return uTxOuts
}

// calculate the balance of the address
func CalTxOuts(address string, b *blockchain) int {
	var balance int
	tx := FilterTxOuts(address, b)
	for _, output := range tx {
		balance += output.Amount
	}
	return balance
}

// get all the transactions inside of blocks
func Transactions(b *blockchain) []*Tx {
	var txs []*Tx
	for _, block := range Blocks(b) {
		txs = append(txs, block.Transactions...)
	}
	return txs
}

// find the transaction inside of blocks
func FindTx(b *blockchain, txId string) *Tx {
	for _, tx := range Transactions(b) {
		if tx.Id == txId {
			return tx
		}
	}
	return nil
}

func GetStatus(b *blockchain, rw http.ResponseWriter) {
	defer b.m.Unlock()
	b.m.Lock()
	err := json.NewEncoder(rw).Encode(b)
	utilities.ErrHandling(err)
}

// Main function: start blockchain
func Blockchain() *blockchain {
	once.Do(func() {
		b = &blockchain{Height: 0}
		bucketData := db.GetBucketData()
		if bucketData == nil {
			b.AddBlock()
		} else {
			utilities.FromByte(b, bucketData)
		}
	})
	return b
}
