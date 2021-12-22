package main

import (
	"encoding/json"
	"net/http"

	"github.com/auturnn/peer-base-nodes/p2p"
	"github.com/gorilla/mux"
)

const port = ":8333"

type addPeerPayload struct {
	Address string `json:"address"`
	Port    string `json:"port"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/peers", func(rw http.ResponseWriter, r *http.Request) {
		json.NewEncoder(rw).Encode(p2p.AllPeers(&p2p.Peers))
	}).Methods("GET")

	r.HandleFunc("/ws", p2p.Upgrade).Methods("GET")
	http.ListenAndServe(port, r)
}
