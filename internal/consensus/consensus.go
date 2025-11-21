package consensus

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/bpjs-hackathon/sehat-chain/internal/p2p"
	"github.com/bpjs-hackathon/sehat-chain/types"
)

type RoundRobin struct {
	ID   string
	Node NodeInterface
	mux  sync.RWMutex

	validators     map[string]string // string id beserta key
	validatorsSort []string          // sorted validator
}

func NewPBFT(id string, node NodeInterface, validators map[string]string) *RoundRobin {
	// sort validator
	validatorsSort := make([]string, 0, len(validators))
	for _, validator := range validators {
		validatorsSort = append(validatorsSort, validator)
	}
	sort.Strings(validatorsSort)

	return &RoundRobin{
		ID:             id,
		Node:           node,
		validators:     validators,
		validatorsSort: validatorsSort,
	}
}

func (r *RoundRobin) StartRound() {
	r.mux.Lock()
	defer r.mux.Unlock()

	if !r.Node.IsValidator() {
		fmt.Printf("node %s cannot start round because it is not a validator\n", r.ID)
		return
	}

	if !r.IsLeader() {
		fmt.Printf("node %s cannot start round because it is not the leader\n", r.ID)
		return
	}

	block := r.Node.CreateBlock()
	// mock signature
	hash := block.HeaderHash()
	block.QC.HeaderHash = hash
	block.QC.Signatures = r.Node.SignData([]byte(hash))

	r.Node.CommitBlock(block)

	// buat payload untuk broadcast
	payload, err := json.Marshal(block)
	if err != nil {
		fmt.Printf("failed to marshal block: %v\n", err)
		return
	}

	blockMsg := p2p.Message{
		SenderID:   r.ID,
		RequestID:  "",
		ResponseID: "",
		Type:       p2p.MsgTypeBlockSend,
		Payload:    payload,
	}

	r.Node.Broadcast(blockMsg)
}

func (r *RoundRobin) HandleIncomingBlock(block types.Block) {
	r.mux.Lock()
	defer r.mux.Unlock()

	latest := r.Node.GetLatestBlock()
	if block.Header.Height != latest.Header.Height+1 {
		fmt.Printf("invalid block height. Expecting %d, got %d\n", latest.Header.Height+1, block.Header.Height)
		return
	}
	if block.Header.PrevHash != latest.HeaderHash() {
		fmt.Printf("invalid previous block hash. Expecting %s, got %s\n", latest.HeaderHash(), block.Header.PrevHash)
		return
	}

	expectedLeader := r.getLeaderForHeight(block.Header.Height)
	if block.Header.ProposerID != expectedLeader {
		fmt.Printf("invalid proposer. Expecting %s, got %s\n", expectedLeader, block.Header.ProposerID)
		return
	}

	fmt.Printf("Incoming block validated, commiting block")
	r.Node.CommitBlock(block)
}

func (r *RoundRobin) IsLeader() bool {
	nextHeight := r.Node.GetLatestBlock().Header.Height + 1
	return r.getLeaderForHeight(nextHeight) == r.ID
}

func (r *RoundRobin) getLeaderForHeight(height uint64) string {
	index := height % uint64(len(r.validatorsSort))
	return r.validatorsSort[index]
}
