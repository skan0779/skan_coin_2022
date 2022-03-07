// Package p2p provides peer to peer functions for skancoin
package p2p

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/skan0779/skan_coin_2022/blockchain"
	"github.com/skan0779/skan_coin_2022/utilities"
)

var upgrader = websocket.Upgrader{}

// ConnNode()'s request를 받으면, upgrade 해줌 || port:3000이 port:4000의 요청 받음
func Upgrade(rw http.ResponseWriter, r *http.Request) {
	address := utilities.Spliter(r.RemoteAddr, ":", 0)
	openPort := r.URL.Query().Get("openPort")
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return address != "" && openPort != ""
	}
	fmt.Printf("\n Received upgrade request from: %s \n", openPort)
	conn, err := upgrader.Upgrade(rw, r, nil)
	utilities.ErrHandling(err)
	startPeer(conn, address, openPort)
}

// port:4000 request upgrade to port:3000 || port:4000가 port:3000에게 연결 요청
func ConnNode(address string, port string, openPort string, check bool) {
	fmt.Printf("\n Start %s connect to: %s \n", openPort, port)
	url := fmt.Sprintf("ws://%s:%s/ws?openPort=%s", address, port, openPort)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	utilities.ErrHandling(err)
	p := startPeer(conn, address, port)
	if check {
		BroadcastNode(p)
		return
	}
	sendLastBlock(p)
}

// broadcast mined new block to all peers
func BroadcastBlock(b *blockchain.Block) {
	for _, peer := range Peers.v {
		sendNewBlock(b, peer)
	}
}

func BroadcastTx(tx *blockchain.Tx) {
	for _, peer := range Peers.v {
		sendNewTx(tx, peer)
	}
}

func BroadcastNode(p *peer) {
	for key, peer := range Peers.v {
		if key != p.key {
			data := fmt.Sprintf("%s:%s", p.key, peer.port)
			sendNewNode(data, peer)
		}
	}
}
