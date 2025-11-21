package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type Peer struct {
	ID      string // identifier setelah melakukan handshake
	Address string

	mux sync.Mutex

	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
}

// loop membaca pesan yang masuk pada koneksi oleh peer
// dan mengirimnya ke sebuah callback
func (p *Peer) readLoop(handler func(message Message)) {
	for {
		var msg Message
		if err := p.decoder.Decode(&msg); err != nil {
			fmt.Printf("peer read error: %v\n", err)
			return
		}
		handler(msg)
	}
}
