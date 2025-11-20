package p2p

import (
	"fmt"
	"time"
)

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
	p2p.pendingResponses[message.RequestID] = responseChannel
	p2p.pendingMux.Unlock()

	// Defer cleanup channel
	defer func() {
		p2p.pendingMux.Lock()
		delete(p2p.pendingResponses, message.RequestID)
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
func (p2p *P2PManager) Broadcast(message Message) {
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
