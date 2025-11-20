// Berisi function untuk memanipulasi chain
package core

import (
	"fmt"
	"sync"
)

type Blockchain struct {
	Blocks []Block
	mux    sync.RWMutex
}

func (bc *Blockchain) AddBlock(block Block) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	lastBlock := bc.Blocks[len(bc.Blocks)-1]
	lastBlockHash := lastBlock.Hash()

	if block.Header.PrevHash != lastBlockHash {
		return fmt.Errorf("invalid previous block hash")
	}

	bc.Blocks = append(bc.Blocks, block)
	return nil
}

func (bc *Blockchain) GetBlock(height uint64) (Block, error) {
	return bc.Blocks[height], nil
}

func (bc *Blockchain) GetLatestBlock() Block {
	return bc.Blocks[len(bc.Blocks)-1]
}
