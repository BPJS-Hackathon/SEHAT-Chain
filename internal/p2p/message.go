package p2p

import (
	"encoding/json"

	"github.com/bpjs-hackathon/sehat-chain/types"
)

// Wrapping message untuk dikirim antar peer
type Message struct {
	SenderID   string          `json:"sender_id"`
	RequestID  string          `json:"request_id"`  // UUID Pesan
	ResponseID string          `json:"response_id"` // RequestID diulang jika balasan
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
}

// Konstanta type pesan p2p
const (
	// HANDSHAKE (START KONEKSI)
	MsgHandshakeReq  = "HANDSHAKE_REQUEST"
	MsgHandshakeResp = "HANDSHAKE_RESPONSE"

	// BLOCKCHAIN DATA
	MsgTypeBlockReq  = "BLOCK_REQUEST"
	MsgTypeBlockSend = "BLOCK_SEND"
	MsgTypePeersReq  = "PEERS_REQUEST"
	MsgTypePeersSend = "PEERS_SEND"

	// CONSENSUS
	MsgTypeTxGossip = "CONSENSUS_TX_GOSSIP" // Node menyebar tx dari frontend/node lain agar semua node menerima tx
	MsgTypeProposal = "CONSENSUS_PROPOSE"   // Leader mengirim block proposal
	MsgTypePrepare  = "CONSENSUS_PREPARE"   // Node mengirim pesan prepare (proposal diterima)
	MsgTypeCommit   = "CONSENSUS_COMMIT"    // Node mengirim pesan commit (siap untuk eksekusi block)
)

// Payloads
// Handshake yang dilakukan antar nodes
// (umumnya ln -> validator & validator -> validator)
// Setelah menerima handshake diteruskan di node dimana ia akan mendeterminasi
// Node ini termasuk ke list validator atau tidak
type HandshakePayload struct {
	NodeID string `json:"node_id"`
	Port   string `json:"port"`
	Secret string `json:"secret"`
}

type BlockRequestPayload struct {
	Height uint64 `json:"height"`
}

type BlockPayload struct {
	LatestHeight uint64      `json:"latest_height"`
	Block        types.Block `json:"block"`
}

type PeerPayload struct {
	Peers map[string]string `json:"peers"` // map id dan address
}

type BlockProposal struct {
	Block types.Block `json:"block"`
}

type VotePayload struct {
	BlockHeight uint64 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	VoteType    string `json:"vote_type"` // "PREPARE" or "COMMIT"
	Signature   string `json:"signature"` // **INI SIGNATURE DARI BLOCK BUKAN DARI PESAN
}

type TxGossipPayload struct {
	types.Transaction
}
