package tcp

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type Peer struct {
	conn    net.Conn
	lock    sync.Mutex // lock to prevent concurrent writes to peer
	encoder *json.Encoder
	decoder *json.Decoder
}

// readLoop continuously reads messages from a peer
func (p *Peer) readLoop(handler func(models.Message)) {
	for {
		var msg models.Message
		if err := p.decoder.Decode(&msg); err != nil {
			fmt.Printf("peer read error: %v\n", err)
			return
		}
		handler(msg)
	}
}
