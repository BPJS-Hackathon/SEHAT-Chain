package core

import (
	"encoding/json"
	"fmt"

	"github.com/bpjs-hackathon/sehat-chain/internal/p2p"
)

func (node *Node) handleIncomingMessage(peer *p2p.Peer, msg p2p.Message) {
	switch msg.Type {
	case p2p.MsgHandshakeReq:
		node.handleHandshakeRequest(peer, msg)
	case p2p.MsgTypeBlockReq:
		node.handleBlockRequest(peer, msg)
	case p2p.MsgTypePeersReq:
		node.handlePeerRequest(peer, msg)
	case p2p.MsgTypeTxGossip:
		node.handleTxGossip(msg)
	case p2p.MsgTypeBlockSend:
		node.handleBlockSend(msg)
	default:
		fmt.Printf("invalid message type")
	}

}

func (node *Node) handleHandshakeRequest(peer *p2p.Peer, message p2p.Message) {
	var handshake p2p.HandshakePayload
	if err := json.Unmarshal(message.Payload, &handshake); err != nil {
		fmt.Printf("invalid message type and actual payload format")
		return
	}

	node.peers[handshake.NodeID] = handshake.Secret

	// Masukkan dalam peer list
	node.P2P.RegisterPeer(peer, handshake.NodeID)

	// Kirimkan pesan balasan ke requester
	respPayload := p2p.HandshakePayload{
		NodeID: node.ID,
		Port:   node.P2P.Port,
		Secret: node.cred.GetSecret(),
	}
	respPayloadRaw, err := json.Marshal(respPayload)
	if err != nil {
		fmt.Printf("handshake resp marshal failed")
	}

	// Wrap kedalam message
	respMessage := p2p.Message{
		SenderID:   node.ID,
		Type:       p2p.MsgHandshakeResp,
		RequestID:  message.RequestID,
		ResponseID: message.ResponseID,
		Payload:    respPayloadRaw,
	}

	node.P2P.Send(handshake.NodeID, respMessage)
}

func (node *Node) handlePeerRequest(peer *p2p.Peer, message p2p.Message) {
	peersList := node.P2P.Peers

	mappedPeers := make(map[string]string)
	for peerID, peer := range peersList {
		mappedPeers[peerID] = peer.Address
	}

	peerResp := p2p.PeerPayload{
		Peers: mappedPeers,
	}
	peerRespRaw, err := json.Marshal(peerResp)
	if err != nil {
		fmt.Printf("peer resp marshal failed")
	}

	respMessage := p2p.Message{
		SenderID:   node.ID,
		RequestID:  message.RequestID,
		ResponseID: message.RequestID,
		Type:       p2p.MsgTypePeersSend,
		Payload:    peerRespRaw,
	}

	node.P2P.Send(peer.ID, respMessage)
}

func (node *Node) handleBlockRequest(peer *p2p.Peer, message p2p.Message) {
	var blockReq p2p.BlockRequestPayload
	if err := json.Unmarshal(message.Payload, &blockReq); err != nil {
		fmt.Printf("invalid message type and actual payload format")
		return
	}

	height := blockReq.Height

	block, err := node.Blockchain.GetBlock(height)
	if err != nil {
		fmt.Printf("cannot handle block request from %s, reason: %s", peer.ID, err)
		return
	}

	blockResp := p2p.BlockPayload{
		LatestHeight: node.Blockchain.GetLatestHeight(),
		Block:        block,
	}
	blockRespRaw, err := json.Marshal(blockResp)
	if err != nil {
		fmt.Printf("block resp marshal failed")
	}

	respMessage := p2p.Message{
		SenderID:   node.ID,
		RequestID:  message.RequestID,
		ResponseID: message.ResponseID,
		Type:       p2p.MsgTypeBlockSend,
		Payload:    blockRespRaw,
	}

	node.P2P.Send(peer.ID, respMessage)
}

func (node *Node) handleTxGossip(message p2p.Message) {
	var txGossip p2p.TxGossipPayload
	if err := json.Unmarshal(message.Payload, &txGossip); err != nil {
		fmt.Print("tx gossip unmarshal failed")
	}

	// Masukkan tx ke mempool
	node.Mempool.AddTransaction(txGossip.Transaction)

	// Broadcast lagi ke list validator
	node.Broadcast(message)

	// Cek jumlah tx
	if node.Mempool.Size() >= MaxBlockTxs {
		node.Consensus.StartRound()
	}
}

func (node *Node) handleBlockSend(message p2p.Message) {
	var blockPayload p2p.BlockPayload
	if err := json.Unmarshal(message.Payload, &blockPayload); err != nil {
		fmt.Print("block payload unmarshal failed")
	}

	node.Consensus.HandleIncomingBlock(blockPayload.Block)
}
