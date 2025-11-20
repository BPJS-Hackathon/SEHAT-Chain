package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

type P2PManager struct {
	ID   string // identifier node kita
	Port string // port yang didengar

	// menyimpan list peer yang terhubung
	Peers    map[string]*Peer // map peer berdasarkan id
	peersMux sync.RWMutex

	// server listener
	listener net.Listener

	// Callback handler (meneruskan pesan ke layer atas)
	messageHandler func(peer *Peer, msg Message)

	// list channel untuk menerima response berdasarkan requestID
	pendingResponses map[string]chan Message
	pendingMux       sync.Mutex
}

// Membuat instance p2p manager
func CreateP2PManager(nodeID string, port string) *P2PManager {
	return &P2PManager{
		ID:               nodeID,
		Port:             port,
		Peers:            make(map[string]*Peer),
		pendingResponses: make(map[string]chan Message),
	}
}

// Membuka dan menerima koneksi p2p
func (p2p *P2PManager) Open() error {
	listener, err := net.Listen("tcp", ":"+p2p.Port)
	if err != nil {
		return err
	}

	p2p.listener = listener

	// jalankan loop untuk menerima request koneksi
	go p2p.acceptLoop()

	// Logging
	log.Printf("TCP P2P Node terbuka pada port %s\n", p2p.Port)
	return nil
}

// Mensubscribe semua pesan yang masuk dari koneksi yang terhubung
// Logika bisnis diatur oleh core blockchain
func (p2p *P2PManager) Subscribe(handler func(peer *Peer, msg Message)) {
	p2p.messageHandler = handler
}

// loop untuk terus menerima koneksi tcp baru
func (p2p *P2PManager) acceptLoop() {
	for {
		conn, err := p2p.listener.Accept()
		if err != nil {
			fmt.Printf("accept error: %v\n", err)
			return
		}

		// Buat peer sementara (belum ada ID karena belum melakukan handshake)
		peer := &Peer{
			conn:    conn,
			encoder: json.NewEncoder(conn),
			decoder: json.NewDecoder(conn),
		}

		// Jalankan readloop untuk peer
		go peer.readLoop(func(message Message) {
			p2p.handleIncomingMessage(peer, message)
		})
	}
}

func (p2p *P2PManager) handleIncomingMessage(peer *Peer, message Message) {
	if message.ResponseID == message.RequestID {
		p2p.pendingMux.Lock()
		channel, exists := p2p.pendingResponses[message.RequestID]
		p2p.pendingMux.Unlock()

		if exists {
			channel <- message
			delete(p2p.pendingResponses, message.RequestID)
			return
		}
	}

	if p2p.messageHandler != nil {
		p2p.messageHandler(peer, message)
	}
}
