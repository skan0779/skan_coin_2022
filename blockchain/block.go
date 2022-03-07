package blockchain

import (
	"strings"
	"time"

	"github.com/skan0779/skan_coin_2022/db"
	"github.com/skan0779/skan_coin_2022/utilities"
)

type Block struct {
	Height       int    `json:"height"`
	Hash         string `json:"hash"`
	PreHash      string `json:"prevHash,omitempty"`
	Difficulty   int    `json:"difficulty"`
	Nonce        int    `json:"nonce"`
	Timestamp    int    `json:"timestamp"`
	Transactions []*Tx  `json:"transactions"`
}

func (b *Block) mine() {
	target := strings.Repeat("0", b.Difficulty)
	for {
		b.Timestamp = int(time.Now().Unix())
		hash := utilities.Hash(b)
		if strings.HasPrefix(hash, target) {
			b.Hash = hash
			break
		} else {
			b.Nonce++
		}
	}
}

func saveBlock(b *Block) {
	db.SaveBlock(b.Hash, utilities.ToByte(b))
}

func createBlock(height int, preHash string, difficulty int) *Block {
	block := &Block{
		Height:     height,
		Hash:       "",
		PreHash:    preHash,
		Difficulty: difficulty,
		Nonce:      0,
	}
	block.mine()
	block.Transactions = Mempool().ConfirmTx()
	saveBlock(block)
	return block
}

func FindBlock(hash string) (*Block, error) {
	data := db.GetBucketBlocks(hash)
	if data == nil {
		return nil, ErrNotFound
	}
	block := &Block{}
	utilities.FromByte(block, data)
	return block, nil
}
