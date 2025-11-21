package consensus

import (
	"github.com/bpjs-hackathon/sehat-chain/internal/p2p"
	"github.com/bpjs-hackathon/sehat-chain/types"
)

// Interface penghubung node dan consensus
type NodeInterface interface {
	Broadcast(message p2p.Message) // mengirim broadcast ke semua peers
	GetLatestBlock() types.Block
	CreateBlock() types.Block      // membuat block proposal
	CommitBlock(block types.Block) // mengcommit block ke blockchain & kirim ke light nodes
	IsValidator() bool
	SignData(data []byte) string // mock signing data
}
