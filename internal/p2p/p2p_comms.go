package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Koneksi awal ke peer (belum melakukan handshake) dan return peer untuk dilakukan handshake
func (p2p *P2PManager) Connect(address string) (*Peer, error) {
	var empty Peer

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return &empty, fmt.Errorf("failed to establish connection to %s, reason %v", address, err)
	}

	peer := &Peer{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
	}

	go peer.readLoop(func(message Message) {
		p2p.handleIncomingMessage(peer, message)
	})

	return peer, nil
}

// Pengiriman pesan one-way (tidak mengharapkan / menunggu response)
func (p2p *P2PManager) Send(peerID string, message Message) error {
	// Cek apakah peer yang ingin kita kirim pesan
	// ada dalam list koneksi
	p2p.peersMux.RLock()
	peer, exists := p2p.Peers[peerID]
	p2p.peersMux.RUnlock()

	if !exists {
		return fmt.Errorf("failed to send p2p message: peer (%s) not found", peerID)
	}

	// Kirim message ke encoder untuk dikirim ke peer
	// Gunakan mutex untuk memastikan write safety saat pengiriman
	peer.mux.Lock()
	defer peer.mux.Unlock()
	return peer.encoder.Encode(message)
}

// Pengiriman pesan two-way (mengirim pesan dan menunggu pesan balasan)
func (p2p *P2PManager) Request(peerID string, message Message, timeout time.Duration) (Message, error) {
	// Menyiapkan 1 channel untuk balasan
	responseChannel := make(chan Message, 1)

	// Registrasi channel ke pending responses
	p2p.pendingMux.Lock()
	p2p.pendingMessages[message.RequestID] = responseChannel
	p2p.pendingMux.Unlock()

	// Defer cleanup channel
	defer func() {
		p2p.pendingMux.Lock()
		delete(p2p.pendingMessages, message.RequestID)
		close(responseChannel)
		p2p.pendingMux.Unlock()
	}()

	// kirim pesan
	if err := p2p.Send(peerID, message); err != nil {
		return Message{}, err
	}

	// tunggu balasan
	select {
	case resp := <-responseChannel:
		return resp, nil
	case <-time.After(timeout):
		return Message{}, fmt.Errorf("p2p request timed out")
	}
}

// Pengiriman pesan one-way ke semua peer terhubung
func (p2p *P2PManager) Broadcast(message Message, peerList []string) {
	// ambil list peers
	p2p.peersMux.RLock()
	peers := p2p.Peers
	p2p.peersMux.RUnlock()

	// kirim pesan ke semua peer
	for _, peer := range peers {
		// goroutine untuk mengirim pesan. tidak menggunakan Send karena
		// ada beberapa checking yang tidak perlu dilakukan disini (performance)
		go func(peer *Peer) {
			peer.mux.Lock()
			defer peer.mux.Unlock()
			if err := peer.encoder.Encode(message); err != nil {
				fmt.Printf("broadcast error to peer (%s): %v\n", peer.ID, err)
			}
		}(peer)
	}
}
