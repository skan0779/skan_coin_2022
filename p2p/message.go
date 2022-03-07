package p2p

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/skan0779/skan_coin_2022/blockchain"
	"github.com/skan0779/skan_coin_2022/utilities"
)

type Message struct {
	Sort    MessageType
	Payload []byte
}
type MessageType int

const (
	MessageLastBlock      MessageType = 0
	MessageBlocksRequest  MessageType = 1
	MessageBlocksResponse MessageType = 2
	MessageNewBlock       MessageType = 3
	MessageNewTx          MessageType = 4
	MessageNewNode        MessageType = 5
)

func sendLastBlock(p *peer) {
	fmt.Printf("Send Last Block to: %s \n", p.key)
	b, err := blockchain.FindBlock(blockchain.Blockchain().PreHash)
	utilities.ErrHandling(err)
	m := getMessage(MessageLastBlock, b)
	p.inbox <- m
}

func getMessage(sort MessageType, payload interface{}) []byte {
	m := Message{
		Sort:    sort,
		Payload: utilities.Json(payload),
	}
	return utilities.Json(m)
}

func requestBlocks(p *peer) {
	fmt.Printf("Request Blocks to: %s \n", p.key)
	m := getMessage(MessageBlocksRequest, nil)
	p.inbox <- m
}

func sendBlocks(p *peer) {
	fmt.Printf("Send Blocks to: %s \n", p.key)
	m := getMessage(MessageBlocksResponse, blockchain.Blocks(blockchain.Blockchain()))
	p.inbox <- m
}

func sendNewBlock(b *blockchain.Block, p *peer) {
	fmt.Printf("Send New Block to: %s \n", p.key)
	m := getMessage(MessageNewBlock, b)
	p.inbox <- m
}

func sendNewTx(tx *blockchain.Tx, p *peer) {
	fmt.Printf("Send New Tx to: %s \n", p.key)
	m := getMessage(MessageNewTx, tx)
	p.inbox <- m
}

func sendNewNode(data string, p *peer) {
	fmt.Printf("Send New Node to: %s \n", p.key)
	m := getMessage(MessageNewNode, data)
	p.inbox <- m
}

func handleMessage(m *Message, p *peer) {
	switch m.Sort {
	case MessageLastBlock:
		fmt.Printf("Received Last Block from: %s \n", p.key)
		var data blockchain.Block
		err := json.Unmarshal(m.Payload, &data)
		utilities.ErrHandling(err)
		block, err := blockchain.FindBlock(blockchain.Blockchain().PreHash)
		utilities.ErrHandling(err)
		if data.Height >= block.Height {
			requestBlocks(p)
		} else {
			sendLastBlock(p)
		}
	case MessageBlocksRequest:
		sendBlocks(p)
	case MessageBlocksResponse:
		fmt.Printf("Received Blocks from: %s \n", p.key)
		var data []*blockchain.Block
		err := json.Unmarshal(m.Payload, &data)
		utilities.ErrHandling(err)
		blockchain.Blockchain().Update(data)
	case MessageNewBlock:
		fmt.Printf("Received New Block from: %s \n", p.key)
		var data *blockchain.Block
		err := json.Unmarshal(m.Payload, &data)
		utilities.ErrHandling(err)
		blockchain.Blockchain().UpdateBlock(data)
	case MessageNewTx:
		fmt.Printf("Received New Tx from: %s \n", p.key)
		var data *blockchain.Tx
		err := json.Unmarshal(m.Payload, &data)
		utilities.ErrHandling(err)
		blockchain.Mempool().UpdateTx(data)
	case MessageNewNode:
		var data string
		err := json.Unmarshal(m.Payload, &data)
		utilities.ErrHandling(err)
		fmt.Printf("Received New Node from: %s \n", data)
		d := strings.Split(data, ":")
		ConnNode(d[0], d[1], d[2], false)
	}
}
