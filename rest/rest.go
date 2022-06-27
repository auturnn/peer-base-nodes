package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/auturnn/peer-base-nodes/blockchain"
	"github.com/auturnn/peer-base-nodes/p2p"
	"github.com/auturnn/peer-base-nodes/utils"
	"github.com/auturnn/peer-base-nodes/wallet"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var port string

type url string

func (u url) MashalText() ([]byte, error) {
	url := fmt.Sprintf("http://localhost%s%s", port, u)
	return []byte(url), nil
}

type urlDescription struct {
	URL         url    `json:"url"`
	Method      string `json:"method"`
	Description string `json:"description"`
	Payload     string `json:"payload,omitempty"`
}

type addBlockBody struct {
	Data string `json:"data"`
}

type errorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

type balanceResponse struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type addTxPayload struct {
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

type myWalletResponse struct {
	Address string `json:"address"`
}

type addPeerPayload struct {
	Address, Port, Wallet string
	Server                bool
}

func documentation(rw http.ResponseWriter, r *http.Request) {
	data := []urlDescription{
		{
			URL:         url("/"),
			Method:      "GET",
			Description: "See Documentation",
		},
		{
			URL:         url("/status"),
			Method:      "GET",
			Description: "See the Status of the Blockchain",
		},
		{
			URL:         url("/blocks"),
			Method:      "GET",
			Description: "See All Blocks",
		},
		{
			URL:         url("/blocks/{hash}"),
			Method:      "GET",
			Description: "See A Block",
		},
		{
			URL:         url("/balance/{address}"),
			Method:      "GET",
			Description: "Get TxOuts for an Address",
		},
		{
			URL:         url("/ws"),
			Method:      "GET",
			Description: "Upgrade to WebSockets",
		},
		{
			URL:         url("/peers"),
			Method:      "GET",
			Description: "Get all connecting Peer's address",
		},
	}
	json.NewEncoder(rw).Encode(data)
}

func getBlocks(rw http.ResponseWriter, r *http.Request) {
	json.NewEncoder(rw).Encode(blockchain.Blocks(blockchain.BlockChain()))
}

func getBlock(rw http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	block, err := blockchain.FindBlock(hash)
	encoder := json.NewEncoder(rw)
	if err == blockchain.ErrNotFound {
		utils.HandleError(encoder.Encode(errorResponse{fmt.Sprint(err)}))
	} else {
		utils.HandleError(encoder.Encode(block))
	}
}

func getStatus(rw http.ResponseWriter, r *http.Request) {
	blockchain.Status(blockchain.BlockChain(), rw)
}

func getBalance(rw http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]
	total := r.URL.Query().Get("total")
	switch total {
	case "true":
		amount := blockchain.BalanceByAddress(address, blockchain.BlockChain())
		utils.HandleError(json.NewEncoder(rw).Encode(balanceResponse{address, amount}))
	default:
		utils.HandleError(json.NewEncoder(rw).Encode(blockchain.UTxOutsByAddress(address, blockchain.BlockChain())))
	}
}

func getMempool(rw http.ResponseWriter, r *http.Request) {
	utils.HandleError(json.NewEncoder(rw).Encode(blockchain.Mempool().Txs))
}

func myWallet(rw http.ResponseWriter, r *http.Request) {
	w := wallet.WalletLayer{}
	address := w.GetAddress()
	json.NewEncoder(rw).Encode(myWalletResponse{Address: address})
}

func getPeers(rw http.ResponseWriter, r *http.Request) {
	json.NewEncoder(rw).Encode(p2p.AllPeers(&p2p.Peers))
}

func postPeer(rw http.ResponseWriter, r *http.Request) {
	var payload addPeerPayload
	json.NewDecoder(r.Body).Decode(&payload)

	p2p.AddPeer(payload.Address, payload.Port, payload.Wallet, port[1:], wallet.WalletLayer{}.GetAddress()[:5], payload.Server)
}

//wallet파일만있으면 자신이 해당 파일을 가지고 그사람인척도 가능.
//그렇기때문에 로그인기능같은 본인인증이 필요함
func Start(cPort int) {
	router := mux.NewRouter()
	router.Use(jsonContentTypeMiddleware, loggerMiddleware)
	router.HandleFunc("/", documentation).Methods("GET")
	router.HandleFunc("/server", p2p.GetServerList).Methods("GET")
	router.HandleFunc("/status", getStatus).Methods("GET")
	router.HandleFunc("/blocks", getBlocks).Methods("GET")
	router.HandleFunc("/blocks/{hash:[a-f0-9]+}", getBlock).Methods("GET")
	router.HandleFunc("/balance/{address}", getBalance).Methods("GET")
	router.HandleFunc("/mempool", getMempool).Methods("GET")
	router.HandleFunc("/wallet", myWallet).Methods("GET")
	router.HandleFunc("/peers", postPeer).Methods("POST")
	router.HandleFunc("/ws", p2p.Upgrade).Methods("GET")
	router.HandleFunc("/peers", getPeers).Methods("GET")
	router.HandleFunc("/health-check", func(rw http.ResponseWriter, r *http.Request) {
		//전 기능 검사후에 healthCheck ok를 true로 바꿀수있도록 개선희망!
		healthCheck := struct {
			Status bool   `json:"status"`
			Msg    string `json:"msg"`
		}{Status: true, Msg: ""}
		json.NewEncoder(rw).Encode(healthCheck)
		rw.WriteHeader(200)
	}).Methods("GET")
	port = fmt.Sprintf(":%d", cPort)
	log.Printf("Listening http://localhost%s\n", port)

	cors := handlers.CORS()(router)
	recovery := handlers.RecoveryHandler()(cors)

	log.Fatal(http.ListenAndServe(port, recovery))
}
