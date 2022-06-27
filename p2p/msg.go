package p2p

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/auturnn/peer-base-nodes/blockchain"
	"github.com/auturnn/peer-base-nodes/utils"
)

type MessageKind int

const (
	MessageNewestBlock MessageKind = iota
	MessageAllBlocksrequest
	MessageAllBlocksResponse
	MessageNewBlockNotify
	MessageNewTxNotify
	MessageNewPeerNotify
)

type Message struct {
	Kind    MessageKind
	Payload []byte
}

func makeMessage(kind MessageKind, payload interface{}) []byte {
	m := Message{
		Kind:    kind,
		Payload: utils.ToJSON(payload),
	}
	return utils.ToJSON(m)
}

func sendNewestBlock(p *peer) {
	log.Printf("Sending newest block to %s\n", p.key)
	b, err := blockchain.FindBlock(blockchain.BlockChain().NewestHash)
	utils.HandleError(err)
	m := makeMessage(MessageNewestBlock, b)
	p.inbox <- m
}

func requestAllBlocks(p *peer) {
	m := makeMessage(MessageAllBlocksrequest, nil)
	p.inbox <- m
}

func sendAllBlocks(p *peer) {
	m := makeMessage(MessageAllBlocksResponse, blockchain.Blocks(blockchain.BlockChain()))
	p.inbox <- m
}

func notifyNewBlock(b *blockchain.Block, p *peer) {
	m := makeMessage(MessageNewBlockNotify, b)
	p.inbox <- m
}

func notifyNewTx(tx *blockchain.Tx, p *peer) {
	m := makeMessage(MessageNewTxNotify, tx)
	p.inbox <- m
}

func notifyNewPeer(addr string, p *peer) {
	m := makeMessage(MessageNewPeerNotify, addr)
	p.inbox <- m
}

func handlerMsg(m *Message, p *peer) {
	switch m.Kind {
	case MessageNewestBlock:
		log.Printf("Received the newest block from %s\n", p.key)

		var payload blockchain.Block
		utils.HandleError(json.Unmarshal(m.Payload, &payload))

		block, err := blockchain.FindBlock(blockchain.BlockChain().NewestHash)
		utils.HandleError(err)

		if payload.Hash == block.Hash && payload.Height == block.Height {
			return
		}

		if payload.Height >= block.Height {
			log.Printf("Requesting all blocks from %s\n", p.key)
			requestAllBlocks(p)
		} else {
			sendNewestBlock(p)
		}

	case MessageAllBlocksrequest:
		log.Printf("%s wants all the blocks.\n", p.key)
		sendAllBlocks(p)

	case MessageAllBlocksResponse:
		log.Printf("Received all the blocks from %s\n", p.key)
		var payload []*blockchain.Block
		utils.HandleError(json.Unmarshal(m.Payload, &payload))
		blockchain.BlockChain().Replace(payload)

	case MessageNewBlockNotify:
		log.Printf("NewBlockNotify!")
		var payload *blockchain.Block
		utils.HandleError(json.Unmarshal(m.Payload, &payload))
		blockchain.BlockChain().AddPeerBlock(payload)

	case MessageNewTxNotify:
		log.Printf("NewTxNotify!")
		var payload *blockchain.Tx
		utils.HandleError(json.Unmarshal(m.Payload, &payload))
		blockchain.Mempool().AddPeerTx(payload)

	case MessageNewPeerNotify:
		var payload string
		// {연결해오는peerAddr : 연결해오는peerPort : 연결해오는peerWallet}
		// :{연결되있는peerAddr: 연결되있는peerPort : 연결되있는peerWallet}
		utils.HandleError(json.Unmarshal(m.Payload, &payload))
		parts := strings.Split(payload, ":")
		server, _ := strconv.ParseBool(parts[5])
		AddPeer(parts[0], parts[1], parts[2], parts[3], parts[4], server)
	}

}
