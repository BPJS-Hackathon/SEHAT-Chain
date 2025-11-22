// Berisi function untuk memanipulasi chain
package core

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bpjs-hackathon/sehat-chain/types"
)

type Blockchain struct {
	Blocks []types.Block
	mux    sync.RWMutex
}

// Buat instance blockchain dengan genesis block yang di hardcode
func InitializeBlockChain() *Blockchain {
	hash := []byte(strings.Repeat("0", 64)) // mengikuti standar All Zeroes sebagai hash untuk genesis block
	hashStr := hex.EncodeToString(hash[:])

	// Buat hardcoded header
	header := types.BlockHeader{
		Height:     0,
		Timestamp:  time.Date(2025, 11, 20, 0, 0, 0, 0, time.UTC).Unix(),
		PrevHash:   hashStr,
		StateRoot:  hashStr,
		TxRoot:     hashStr,
		ProposerID: "SYSTEM_INIT",
	}

	// Buat block
	block := types.Block{
		Header:       header,
		Transactions: []types.Transaction{},
		QC:           types.QuorumCertificate{},
	}

	// QC kosong
	block.QC = types.QuorumCertificate{
		HeaderHash: block.HeaderHash(),
		Signatures: "",
	}

	blockchain := Blockchain{
		Blocks: []types.Block{block},
	}

	return &blockchain
}

// Tambah block ke chain (fixed block setelah Commit consensus)
func (bc *Blockchain) AddBlock(block types.Block) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	lastBlock := bc.Blocks[len(bc.Blocks)-1]

	if block.Header.Height != lastBlock.Header.Height+1 {
		return fmt.Errorf("invalid block height (after commit). Expecting %d, got %d", lastBlock.Header.Height+1, block.Header.Height)
	}

	if block.Header.PrevHash != lastBlock.HeaderHash() {
		return fmt.Errorf("previous block hash did not match")
	}

	txSeen := make(map[string]bool)
	for _, existingBlock := range bc.Blocks {
		for _, tx := range existingBlock.Transactions {
			txSeen[tx.ID] = true
		}
	}

	bc.Blocks = append(bc.Blocks, block)
	return nil
}

// Mengambil block dari height yang spesifik
func (bc *Blockchain) GetBlock(height uint64) (types.Block, error) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	if height >= uint64(len(bc.Blocks)) {
		return types.Block{}, fmt.Errorf("requested height is higher than current stored")
	}

	block := bc.Blocks[height]
	return block, nil
}

func (bc *Blockchain) GetLatestHeight() uint64 {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	return bc.Blocks[len(bc.Blocks)-1].Header.Height
}

func (bc *Blockchain) GetLatestBlock() types.Block {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.Blocks[len(bc.Blocks)-1]
}
