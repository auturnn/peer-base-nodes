package blockchain

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/auturnn/peer-base-nodes/db"
	"github.com/auturnn/peer-base-nodes/utils"
)

const (
	defaultDiffculty   int = 2
	difficultyInterval int = 5
	blockInterval      int = 2
	allowedRange       int = 2
)

type blockchain struct {
	NewestHash        string     `json:"newestHash"`
	Height            int        `json:"height"`
	CurrentDifficulty int        `json:"currentDifficulty"`
	m                 sync.Mutex `json:"-"`
}

type storage interface {
	FindBlock(hash string) []byte
	LoadChain() []byte
	SaveBlock(hash string, data []byte)
	SaveChain(data []byte)
	DeleteAllBlocks()
}

var bc *blockchain
var once sync.Once
var dbStorage storage = db.DB{}

func Txs(b *blockchain) []*Tx {
	var txs []*Tx
	for _, block := range Blocks(b) {
		txs = append(txs, block.Transactions...)
	}
	return txs
}

func FindTx(bc *blockchain, targetID string) *Tx {
	for _, tx := range Txs(bc) {
		if tx.ID == targetID {
			return tx
		}
	}
	return nil
}

func recalculrateDifficulty(bc *blockchain) int {
	allBlocks := Blocks(bc)
	newestBlock := allBlocks[0]
	lastRecalculratedBlock := allBlocks[difficultyInterval-1]
	//actualTime : 현재 생성되는 블럭의 생성주기
	actualTime := (newestBlock.Timestamp / 60) - (lastRecalculratedBlock.Timestamp / 60)
	//expectedTime : 의도한 블럭생성주기
	expectedTime := difficultyInterval * blockInterval
	if actualTime < (expectedTime - allowedRange) {
		return bc.CurrentDifficulty + 1
	} else if actualTime > (expectedTime + allowedRange) {
		return bc.CurrentDifficulty - 1
	}
	return bc.CurrentDifficulty
}

func getDifficulty(bc *blockchain) int {
	if bc.Height == 0 {
		return defaultDiffculty
	} else if bc.Height%difficultyInterval == 0 {
		//recalculrate the difficulty
		return recalculrateDifficulty(bc)
	} else {
		return bc.CurrentDifficulty
	}
}

func persistBlockchain(bc *blockchain) {
	dbStorage.SaveChain(utils.ToBytes(bc))
}

func (bc *blockchain) AddBlock() *Block {
	block := createBlock(bc.NewestHash, bc.Height+1, getDifficulty(bc))
	bc.NewestHash = block.Hash
	bc.Height = block.Height
	bc.CurrentDifficulty = block.Difficulty
	persistBlockchain(bc)
	return block
}

func (bc *blockchain) restore(data []byte) {
	utils.FromBytes(bc, data)
}

func UTxOutsByAddress(address string, bc *blockchain) []*UTxOut {
	var uTxOuts []*UTxOut
	creatorTxs := make(map[string]bool)

	for _, block := range Blocks(bc) {
		for _, tx := range block.Transactions {
			for _, input := range tx.TxIns {
				if input.Signature == "COINBASE" {
					break
				}
				if FindTx(bc, input.TxID).TxOuts[input.Index].Address == address {
					creatorTxs[input.TxID] = true
				}
			}
			for index, output := range tx.TxOuts {
				if output.Address == address {
					if _, ok := creatorTxs[tx.ID]; !ok {
						uTxOut := &UTxOut{tx.ID, index, output.Amount}
						if !isOnMempool(uTxOut) {
							uTxOuts = append(uTxOuts, uTxOut)
						}
					}
				}
			}
		}
	}
	return uTxOuts
}

func BalanceByAddress(address string, bc *blockchain) (balance int) {
	txOuts := UTxOutsByAddress(address, bc)
	for _, txOut := range txOuts {
		balance += txOut.Amount
	}
	return balance
}

func BlockChain() *blockchain {
	once.Do(func() {
		bc = &blockchain{Height: 0}
		checkpoint := dbStorage.LoadChain()
		if checkpoint != nil {
			bc.restore(checkpoint)
		}
	})
	return bc
}

func Blocks(bc *blockchain) (blocks []*Block) {
	bc.m.Lock()
	defer bc.m.Unlock()
	hashCursor := bc.NewestHash
	for {
		block, err := FindBlock(hashCursor)
		utils.HandleError(err)
		blocks = append(blocks, block)
		if block.PrevHash != "" {
			hashCursor = block.PrevHash
		} else {
			break
		}
	}
	return blocks
}

func (bc *blockchain) Replace(newBlocks []*Block) {
	bc.m.Lock()
	defer bc.m.Unlock()
	bc.CurrentDifficulty = newBlocks[0].Difficulty
	bc.Height = len(newBlocks)
	bc.NewestHash = newBlocks[0].Hash
	persistBlockchain(bc)
	dbStorage.DeleteAllBlocks()
	for _, block := range newBlocks {
		persistBlock(block)
	}
}

func Status(bc *blockchain, rw http.ResponseWriter) {
	bc.m.Lock()
	defer bc.m.Unlock()

	utils.HandleError(json.NewEncoder(rw).Encode(bc))
}

func (bc *blockchain) AddPeerBlock(newBlock *Block) {
	bc.m.Lock()
	mp.m.Lock()
	defer bc.m.Unlock()
	defer mp.m.Unlock()

	bc.Height++
	bc.CurrentDifficulty = newBlock.Difficulty
	bc.NewestHash = newBlock.Hash

	persistBlockchain(bc)
	persistBlock(newBlock)

	//mempool
	for _, tx := range newBlock.Transactions {
		if _, ok := mp.Txs[tx.ID]; ok {
			delete(mp.Txs, tx.ID)
		}
	}
}
