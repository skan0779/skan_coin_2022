package p2p

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type peers struct {
	v map[string]*peer
	m sync.Mutex
}
type peer struct {
	conn    *websocket.Conn
	inbox   chan []byte
	key     string
	address string
	port    string
}

var Peers peers = peers{
	v: make(map[string]*peer),
}

func (p *peer) read() {
	defer p.close()
	for {
		m := Message{}
		err := p.conn.ReadJSON(&m)
		if err != nil {
			break
		}
		handleMessage(&m, p)
	}
}

func (p *peer) write() {
	defer p.close()
	for {
		message, ok := <-p.inbox
		if !ok {
			break
		}
		err := p.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

func (p *peer) close() {
	defer Peers.m.Unlock()
	Peers.m.Lock()
	p.conn.Close()
	delete(Peers.v, p.key)
}

func GetPeers(p *peers) []string {
	defer p.m.Unlock()
	p.m.Lock()
	var peers []string
	for key := range p.v {
		peers = append(peers, key)
	}
	return peers
}

func startPeer(conn *websocket.Conn, address string, port string) *peer {
	defer Peers.m.Unlock()
	Peers.m.Lock()
	key := fmt.Sprintf("%s:%s", address, port)
	p := &peer{
		conn:    conn,
		inbox:   make(chan []byte),
		key:     key,
		address: address,
		port:    port,
	}
	Peers.v[key] = p
	go p.read()
	go p.write()
	return p
}
