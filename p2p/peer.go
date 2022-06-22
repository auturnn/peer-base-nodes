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

var Peers peers = peers{
	v: make(map[string]*peer),
}

type peer struct {
	key   string
	addr  string
	port  string
	wAddr string
	conn  *websocket.Conn
	inbox chan []byte
}

func AllPeers(p *peers) []string {
	p.m.Lock()
	defer p.m.Unlock()

	var keys []string
	for key := range p.v {
		keys = append(keys, key)
	}

	return keys
}

func (p *peer) close() {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	p.conn.Close()
	delete(Peers.v, p.key)
}

func (p *peer) read() {
	defer p.close()
	//delete peer in case of error
	for {
		m := Message{}
		err := p.conn.ReadJSON(&m) //block
		if err != nil {
			break
		}
		handlerMsg(&m, p)
	}
}

func (p *peer) write() {
	defer p.close()
	for {
		msg, ok := <-p.inbox //block
		if !ok {
			break
		}
		p.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func initPeer(conn *websocket.Conn, addr, port, wAddr string) *peer {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	key := fmt.Sprintf("%s:%s:%s", addr, port, wAddr)
	p := &peer{
		addr:  addr,
		port:  port,
		key:   key,
		wAddr: wAddr,
		conn:  conn,
		inbox: make(chan []byte),
	}

	go p.read()
	go p.write()
	Peers.v[key] = p
	return p
}
