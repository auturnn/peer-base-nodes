package blockchain

import (
	"errors"
	"strings"
	"time"

	"github.com/auturnn/peer-base-nodes/utils"
)

type Block struct {
	Hash         string `json:"hash"`
	PrevHash     string `json:"prevHash"`
	Height       int    `json:"height"`
	Difficulty   int    `json:"difficulty"`
	Nonce        int    `json:"nonce"`
	Timestamp    int    `json:"timestamp"`
	Miner        string `json:"miner"`
	Transactions []*Tx  `json:"transactions"`
}

func persistBlock(b *Block) {
	dbStorage.SaveBlock(b.Hash, utils.ToBytes(b))
}

func (b *Block) restore(data []byte) {
	utils.FromBytes(b, data)
}

var ErrNotFound = errors.New("block not found")

func FindBlock(hash string) (*Block, error) {
	blockBytes := dbStorage.FindBlock(hash)
	if blockBytes == nil {
		return nil, ErrNotFound
	}
	block := &Block{}
	block.restore(blockBytes)
	return block, nil
}

func (b *Block) mine() {
	target := strings.Repeat("0", b.Difficulty)
	for {
		b.Timestamp = int(time.Now().Unix())
		b.Miner = w.GetAddress()
		hash := utils.Hash(b)
		// fmt.Printf("Hash:=%s\nTarget:=%s\nNonce:=%d\n\n", hash, target, b.Nonce)
		if !strings.HasPrefix(hash, target) {
			b.Nonce++
		} else {
			b.Hash = hash
			break
		}
	}
}

func createBlock(prevHash string, height, diff int) *Block {
	block := &Block{
		PrevHash:   prevHash,
		Hash:       "",
		Height:     height,
		Difficulty: diff,
		Nonce:      0,
	}
	block.Transactions = Mempool().TxToConfirm()
	//mining이 언제 끝날지 모르기 때문에 끝난후에 이떄 추가
	block.mine()
	persistBlock(block)
	return block
}
