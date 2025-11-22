// Berisi komunikasi antar komponen block chain
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bpjs-hackathon/sehat-chain/internal/api"
	"github.com/bpjs-hackathon/sehat-chain/internal/consensus"
	"github.com/bpjs-hackathon/sehat-chain/internal/p2p"
	smartcontract "github.com/bpjs-hackathon/sehat-chain/internal/smart_contract"
	"github.com/bpjs-hackathon/sehat-chain/internal/state"
	"github.com/bpjs-hackathon/sehat-chain/types"
	"github.com/bpjs-hackathon/sehat-chain/utils"
	"github.com/google/uuid"
)

const (
	MaxBlockTxs = 1
)

type Node struct {
	// Identitas node
	ID          string
	cred        *utils.CryptoCred
	isValidator bool

	// NB: Untuk MVP Proto kita tidak menggunakan keypairs melainkan hanya string secret (mock)
	// List peers map[id]secret
	validators map[string]types.ValidatorConfig
	peers      map[string]string

	// Data
	Blockchain *Blockchain
	WorldState *state.WorldState
	Executor   *smartcontract.Executor
	P2P        *p2p.P2PManager
	Consensus  *consensus.RoundRobin

	// Pool
	txPool []types.Transaction
	txMap  map[string]types.Transaction

	// API
	Server *api.Server

	mux sync.RWMutex
}

func CreateNode(ID string, secret string, port string, APIPort string, validators []types.ValidatorConfig) *Node {
	// 1. Convert Slice to Map untuk lookup cepat
	validatorsMap := make(map[string]types.ValidatorConfig)
	for _, v := range validators {
		validatorsMap[v.ID] = v
	}
	_, isValidator := validatorsMap[ID]

	p2pMan := p2p.CreateP2PManager(ID, port)

	blockchain := InitializeBlockChain()
	ws := state.CreateWorldState()

	executor := smartcontract.NewExecutor(ws)

	node := Node{
		ID:          ID,
		isValidator: isValidator,
		validators:  validatorsMap,
		peers:       make(map[string]string),
		cred:        utils.NewCred(secret),
		Blockchain:  blockchain,
		WorldState:  ws,
		Executor:    executor,
		P2P:         p2pMan,
		txPool:      make([]types.Transaction, 0),
		txMap:       make(map[string]types.Transaction),
	}

	node.Consensus = consensus.NewRoundRobin(ID, &node, validatorsMap)
	node.P2P.Subscribe(node.handleIncomingMessage)

	handler := api.CreateAPIHandler()
	handler.AddEndpoint("POST /api/rekam_medis/fk1", node.handleFK1RekamMedisPost)
	handler.AddEndpoint("POST /api/rekam_medis/fk2", node.handleFK2RekamMedisPost)
	handler.AddEndpoint("POST /api/claim", node.handleClaimExecute)
	handler.AddEndpoint("GET /api/total_block", node.handleBlockTotalReq)
	handler.AddEndpoint("GET /api/block/{height}", node.handleAPIBlockRequest)

	server := api.CreateServer(handler, APIPort)
	node.Server = server

	return &node
}

func (node *Node) Start() {
	if err := node.P2P.Open(); err != nil {
		panic(err)
	}

	go node.Server.Run()

	//node.ConnectToNetwork()
	// node.EfficientConnectToNetwork()
	node.RobustConnectToNetwork()
}

func (node *Node) ConnectToNetwork() {
	time.Sleep(time.Second * 2)

	// connect to fixed validators
	var wg sync.WaitGroup

	for _, validator := range node.validators {
		if validator.ID == node.ID {
			continue
		}

		wg.Add(1)
		go func(address string) {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				if err := node.startHandshake(address); err == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}
		}(validator.Address)
	}

	// request block terbaru (untuk melihat ketinggalan)
	reqPayload := p2p.BlockRequestPayload{
		Height: node.Blockchain.GetLatestHeight() + 1,
	}

	reqPayloadRaw, _ := json.Marshal(reqPayload)

	reqMessage := p2p.Message{
		SenderID:  node.ID,
		RequestID: uuid.NewString(),
		Type:      p2p.MsgTypeBlockReq,
		Payload:   reqPayloadRaw,
	}

	//var latestHeight uint64
	for k := range node.validators {
		if k == node.ID {
			continue
		}

		resp, err := node.P2P.Request(k, reqMessage, time.Second*5)
		if err != nil {
			continue
		}

		var blockPayload p2p.BlockPayload
		if err := json.Unmarshal(resp.Payload, &blockPayload); err != nil {
			continue
		}

		//latestHeight = blockPayload.LatestHeight
		break
	}

	// sinkronisasi network
	node.triggerSync()
}

func (node *Node) EfficientConnectToNetwork() {
	time.Sleep(time.Second * 1)

	var wg sync.WaitGroup

	for _, validator := range node.validators {
		if validator.ID == node.ID {
			continue
		}

		// Only connect if our ID is "less than" theirs (prevents both sides connecting)
		if node.ID >= validator.ID {
			continue // Let the other node initiate
		}

		wg.Add(1)
		go func(address string) {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				if err := node.startHandshake(address); err == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}
		}(validator.Address)
	}

	wg.Wait() // Wait for all handshakes
}

func (node *Node) RobustConnectToNetwork() {
	time.Sleep(time.Second * 1)

	var wg sync.WaitGroup
	connectionAttempts := make(map[string]int)
	var connMux sync.Mutex

	for _, validator := range node.validators {
		if validator.ID == node.ID {
			continue
		}

		wg.Add(1)
		go func(validatorID, address string) {
			defer wg.Done()

			// Wait a bit if we're the "higher" ID to let the other side connect first
			if node.ID > validatorID {
				time.Sleep(2 * time.Second)

				// Check if peer already connected to us
				node.P2P.PeersMux.RLock()
				_, alreadyConnected := node.P2P.Peers[validatorID]
				node.P2P.PeersMux.RUnlock()

				if alreadyConnected {
					fmt.Printf("✅ Peer %s already connected (initiated by them)\n", validatorID)
					return
				}
			}

			// Attempt connection
			for i := 0; i < 10; i++ {
				// Check again if peer connected while we were retrying
				node.P2P.PeersMux.RLock()
				_, alreadyConnected := node.P2P.Peers[validatorID]
				node.P2P.PeersMux.RUnlock()

				if alreadyConnected {
					fmt.Printf("✅ Peer %s connected during retry\n", validatorID)
					return
				}

				if err := node.startHandshake(address); err == nil {
					connMux.Lock()
					connectionAttempts[validatorID] = i + 1
					connMux.Unlock()
					fmt.Printf("✅ Successfully connected to %s (attempt %d)\n", validatorID, i+1)
					return
				}

				fmt.Printf("⚠️ Connection attempt %d to %s failed, retrying...\n", i+1, validatorID)
				time.Sleep(1 * time.Second)
			}

			fmt.Printf("❌ Failed to connect to %s after 3 attempts\n", validatorID)
		}(validator.ID, validator.Address)
	}

	wg.Wait()

	// Log final status
	node.P2P.PeersMux.RLock()
	actualPeerCount := len(node.P2P.Peers)
	node.P2P.PeersMux.RUnlock()

	fmt.Printf("Connecting finished with final peer count: %d\n",
		actualPeerCount)

	// Trigger sync
	time.Sleep(500 * time.Millisecond)
	node.triggerSync()
}

func (node *Node) triggerSync() {

	currentHeight := node.Blockchain.GetLatestHeight()
	reqPayload := p2p.BlockRequestPayload{
		Height: currentHeight + 1, // Minta blok selanjutnya (dummy request untuk dapat metadata height)
	}
	// Note: Sebaiknya ada message tipe CHAIN_STATE_REQ untuk minta height saja,
	// tapi pakai BlockReq juga bisa jika responnya mengandung LatestHeight.

	reqPayloadRaw, _ := json.Marshal(reqPayload)
	reqMessage := p2p.Message{
		SenderID:  node.ID,
		RequestID: uuid.NewString(),
		Type:      p2p.MsgTypeBlockReq,
		Payload:   reqPayloadRaw,
	}

	var maxNetworkHeight uint64 = 0

	// Tanya ke semua validator yang terhubung
	for id := range node.validators {
		if id == node.ID {
			continue
		}

		// Gunakan Request dengan timeout pendek
		resp, err := node.P2P.Request(id, reqMessage, 2*time.Second)
		if err != nil {
			fmt.Printf("request error : %v", err)
			continue
		}

		var blockPayload p2p.BlockPayload
		if err := json.Unmarshal(resp.Payload, &blockPayload); err == nil {
			if blockPayload.LatestHeight > maxNetworkHeight {
				maxNetworkHeight = blockPayload.LatestHeight
			}
		}
	}

	if maxNetworkHeight > currentHeight {
		fmt.Printf("📉 Node behind! Current: %d, Network: %d. Starting sync...\n", currentHeight, maxNetworkHeight)
		go node.syncChain(maxNetworkHeight)
	} else {
		fmt.Println("✅ Node up-to-date.")
	}
}

func (node *Node) syncChain(targetHeight uint64) {
	fmt.Printf("🔄 Syncing chain up to height %d\n", targetHeight)

	for {
		currentHeight := node.Blockchain.GetLatestHeight()
		if currentHeight >= targetHeight {
			fmt.Println("✅ Sync Complete!")
			return
		}

		nextHeight := currentHeight + 1
		fmt.Printf("⬇️ Requesting Block #%d...\n", nextHeight)

		reqPayload := p2p.BlockRequestPayload{Height: nextHeight}
		reqPayloadRaw, _ := json.Marshal(reqPayload)

		reqMessage := p2p.Message{
			SenderID:  node.ID,
			RequestID: uuid.NewString(),
			Type:      p2p.MsgTypeBlockReq,
			Payload:   reqPayloadRaw,
		}

		blockReceived := false

		// Coba minta ke validator satu per satu sampai dapat
		for id := range node.validators {
			if id == node.ID {
				continue
			}

			resp, err := node.P2P.Request(id, reqMessage, 2*time.Second)
			if err != nil {
				continue
			}

			var blockPayload p2p.BlockPayload
			if err := json.Unmarshal(resp.Payload, &blockPayload); err != nil {
				continue
			}

			// Validasi blok yang diterima sebelum commit
			newBlock := blockPayload.Block
			if newBlock.Header.Height == nextHeight {
				// CommitBlock sudah handle validasi lanjutan & insert DB
				node.CommitBlock(newBlock)
				blockReceived = true

				// Update target jika network tumbuh saat kita sync
				if blockPayload.LatestHeight > targetHeight {
					targetHeight = blockPayload.LatestHeight
				}

				break // Lanjut ke blok berikutnya
			}
		}

		if !blockReceived {
			fmt.Printf("⚠️ Failed to fetch Block #%d from any peer. Retrying...\n", nextHeight)
			time.Sleep(2 * time.Second)
		} else {
			// Istirahat sebentar biar gak spam network
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (node *Node) startHandshake(address string) error {
	// Buat hubungan ke node
	peer, err := node.P2P.Connect(address)
	if err != nil {
		return err
	}

	// Buat message handshake
	handshake := p2p.HandshakePayload{
		NodeID: node.ID,
		Port:   node.P2P.Port,
		Secret: node.cred.GetSecret(),
	}

	handshakeJson, _ := json.Marshal(handshake)

	// wrap message
	message := p2p.Message{
		SenderID:  node.ID,
		RequestID: uuid.NewString(),
		Type:      p2p.MsgHandshakeReq,
		Payload:   handshakeJson,
	}

	// Kirim message dan tunggu balasan (blocking)
	responseMessage, err := node.P2P.Request(address, message, time.Second*5)
	if err != nil {
		return err
	}

	// Parsing balasan
	var respPayload p2p.HandshakePayload
	if err := json.Unmarshal(responseMessage.Payload, &respPayload); err != nil {
		return err
	}

	// Register peer
	node.mux.Lock()
	node.peers[respPayload.NodeID] = respPayload.Secret
	node.mux.Unlock()

	node.P2P.RemovePeer(address)
	node.P2P.RegisterPeer(peer, respPayload.NodeID)

	fmt.Printf("connecting & handshake to %s got id as such %s", address, respPayload.NodeID)

	return nil
}

func (node *Node) Broadcast(message p2p.Message) {
	nodeIDs := make([]string, 0, len(node.peers))
	for k := range node.peers {
		nodeIDs = append(nodeIDs, k)
	}
	node.P2P.Broadcast(message, nodeIDs)
}

func (node *Node) GetLatestBlock() types.Block {
	return node.Blockchain.GetLatestBlock()
}

func (node *Node) CommitBlock(block types.Block) {
	node.mux.Lock()
	node.Blockchain.AddBlock(block)

	node.Executor.ApplyBlock(block)

	node.RemoveTxsFromPool(len(block.Transactions))
	node.mux.Unlock()

	blockPayload := p2p.BlockPayload{
		LatestHeight: node.Blockchain.GetLatestHeight(),
		Block:        block,
	}
	blockPayloadRaw, _ := json.Marshal(blockPayload)
	node.Broadcast(p2p.Message{
		SenderID:   node.ID,
		RequestID:  uuid.NewString(),
		ResponseID: "",
		Type:       p2p.MsgTypeBlockSend,
		Payload:    blockPayloadRaw,
	})
}

func (node *Node) IsValidator() bool {
	return node.isValidator
}

// Sign dan return hex encoded signature
func (node *Node) SignData(data []byte) string {
	return node.cred.Sign(data)
}

func (node *Node) CreateBlock() types.Block {
	txCount := min(len(node.txPool), MaxBlockTxs)
	txs := node.txPool[:txCount]

	prevBlock := node.Blockchain.GetLatestBlock()

	txRoot := calculateTxRoot(txs)
	stateRoot := node.WorldState.CalculateHash()

	header := types.BlockHeader{
		Height:     prevBlock.Header.Height + 1,
		Timestamp:  time.Now().Unix(),
		PrevHash:   prevBlock.HeaderHash(),
		StateRoot:  stateRoot,
		TxRoot:     txRoot,
		ProposerID: node.ID,
	}

	return types.Block{
		Header:       header,
		Transactions: txs,
		QC:           types.QuorumCertificate{},
	}
}

func calculateTxRoot(txs []types.Transaction) string {
	if len(txs) == 0 {
		return strings.Repeat("0", 64)
	}
	data := ""
	for _, tx := range txs {
		data += tx.ID
	}
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Helper submit tx ke network
func (node *Node) submitTransactionToNetwork(tx types.Transaction) {
	node.AddTxToPool(tx)

	payload := p2p.TxGossipPayload{
		Transaction: tx,
	}
	payloadJson, _ := json.Marshal(payload)

	msg := p2p.Message{
		SenderID:  node.ID,
		RequestID: uuid.NewString(),
		Type:      p2p.MsgTypeTxGossip,
		Payload:   payloadJson,
	}

	// 3. Broadcast to peers
	node.Broadcast(msg)
}

// Add tx to pool
func (node *Node) AddTxToPool(tx types.Transaction) bool {
	_, exists := node.txMap[tx.ID]
	if exists {
		fmt.Print("skipping tx because it already exist on the pool\n")
		return false
	}
	node.txPool = append(node.txPool, tx)
	node.txMap[tx.ID] = tx
	fmt.Print("added 1 tx to the pool\n")
	return true
}

// Remove tx from pool
func (node *Node) RemoveTxsFromPool(count int) bool {
	if len(node.txPool) < count {
		return false
	}

	for i := 0; i < count; i++ {
		delete(node.txMap, node.txPool[i].ID)
	}
	node.txPool = node.txPool[count:]
	return true
}
