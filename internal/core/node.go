// Berisi komunikasi antar komponen block chain
package core

import (
	"github.com/bpjs-hackathon/sehat-chain/internal/mempool"
	smartcontract "github.com/bpjs-hackathon/sehat-chain/internal/smart_contract"
	"github.com/bpjs-hackathon/sehat-chain/types"
)

type Node struct {
	Blockchain            *Blockchain
	Mempool               *mempool.MemPool
	SmartContractExecutor *smartcontract.Executor
	P2P                   *P2PManager
}

func (node *Node) SubmitTx(tx types.Transaction) {
	// validasi signature tx
	// Masukkan ke mempool
	// Broadcast ke peers
}

func (node *Node) CommitBlock(block Block) {
	// jalankan smart contract executor
	// simpan ke blockchain
	// trigger update ke sql jika node adalah bpjs server
}
