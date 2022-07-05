package p2p

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/auturnn/peer-base-nodes/utils"
	"github.com/auturnn/peer-base-nodes/wallet"
	"github.com/gorilla/websocket"
	log "github.com/kataras/golog"
)

var logf = log.Logf
var upgrader = websocket.Upgrader{}
var myWalletAddr = wallet.WalletLayer{}.GetAddress()[:5]

func Upgrade(rw http.ResponseWriter, r *http.Request) {
	//port :3000 will upgrade the request from :4000
	//AddPeer에서는 기존의 peer에 저장되있던 node들의 정보를 port와 waddr 쿼리로 보내지만
	//Upgrade를 받는 쪽에서는 해당 노드들이 새로 연결을 요청하는 쪽이기 때문에 newPeer가 된다.
	if r.URL.Query().Get("nwddr") != myWalletAddr {
		return
	}

	newPeerPort := r.URL.Query().Get("port")
	newPeerWddr := r.URL.Query().Get("wddr")
	newPeerServerTF := r.URL.Query().Get("server")
	ip := utils.Splitter(r.RemoteAddr, ":", 0)

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return newPeerPort != "" && ip != ""
	}
	logf(log.InfoLevel, "%s:%s:%s wants an upgrade", ip, newPeerPort, newPeerWddr)

	conn, err := upgrader.Upgrade(rw, r, nil)
	utils.HandleError(err)
	server, _ := strconv.ParseBool(newPeerServerTF)
	p := initPeer(conn, ip, newPeerPort, newPeerWddr, server)
	broadcastNewPeer(p)
}

func AddPeer(newPeerAddr, newPeerPort, newPeerWAddr, existPeerPort, existPeerWAddr string, server bool) {
	logf(log.InfoLevel, "%s:%s:%s wants to connect to port %s:%s", newPeerAddr, newPeerPort, newPeerWAddr, existPeerPort, existPeerWAddr)
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%s/ws?nwddr=%s&port=%s&wddr=%s&server=%t", newPeerAddr, newPeerPort, newPeerWAddr, existPeerPort, existPeerWAddr, server), nil)
	utils.HandleError(err)
	p := initPeer(conn, newPeerAddr, newPeerPort, newPeerWAddr, server)
	sendNewestBlock(p)
}

func broadcastNewPeer(newPeer *peer) {
	for key, p := range Peers.v {
		if key != newPeer.key {
			//payload = {newPeerAddr}:{newPeerPort}:{newPeerWalletAddr}:{existPeerPort}:{existPeerWalletAddr}:{serverBool}
			payload := fmt.Sprintf("%s:%s:%s:%t", newPeer.key, p.port, p.wddr, p.server)
			notifyNewPeer(payload, p)
		}
	}
}

func GetServerList(rw http.ResponseWriter, r *http.Request) {
	serverList := []string{}
	for key, p := range Peers.v {
		if p.server {
			serverList = append(serverList, key)
		}
	}
	json.NewEncoder(rw).Encode(serverList)
}
