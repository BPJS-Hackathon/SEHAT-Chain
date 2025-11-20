package p2p

// Wrapping message untuk dikirim antar peer
type Message struct {
	SenderID   string `json:"sender_id"`
	RequestID  string `json:"request_id"`  // UUID Pesan
	ResponseID string `json:"response_id"` // RequestID diulang jika balasan
	Type       string `json:"type"`
	Payload    string `json:"payload"`
}

// Konstanta type pesan p2p
const (
	// HANDSHAKE (START KONEKSI)
	MsgHandshakeReq  = "HANDSHAKE_REQUEST"
	MsgHandshakeResp = "HANDSHAKE_RESPONSE"

	// BLOCKCHAIN DATA
	MsgTypeBlockReq       = "BLOCK_REQUEST"
	MsgTypeBlockSend      = "BLOCK_SEND"
	MsgTypePeerReq        = "PEER_REQUEST"
	MsgTypePeerSend       = "PEER_SEND"
	MsgTypeChainStateReq  = "CHAIN_STATE_REQUEST"
	MsgTypeChainStateSend = "CHAIN_STATE_SEND"

	// CONSENSUS
	MsgTypeTxGossip = "CONSENSUS_TX_GOSSIP" // Node menyebar tx dari frontend/node lain agar semua node menerima tx
	MsgTypeProposal = "CONSENSUS_PROPOSE"   // Leader mengirim block proposal
	MsgTypePrepare  = "CONSENSUS_PREPARE"   // Node mengirim pesan prepare (proposal diterima)
	MsgTypeCommit   = "CONSENSUS_COMMIT"    // Node mengirim pesan commit (siap untuk eksekusi block)
)
