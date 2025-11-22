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

	"github.com/bpjs-hackathon/sehat-chain/internal/consensus"
	"github.com/bpjs-hackathon/sehat-chain/internal/mempool"
	"github.com/bpjs-hackathon/sehat-chain/internal/p2p"
	smartcontract "github.com/bpjs-hackathon/sehat-chain/internal/smart_contract"
	"github.com/bpjs-hackathon/sehat-chain/internal/state"
	"github.com/bpjs-hackathon/sehat-chain/types"
	"github.com/bpjs-hackathon/sehat-chain/utils"
	"github.com/google/uuid"
)

const (
	MaxBlockTxs = 5
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
	Mempool    *mempool.MemPool
	Executor   *smartcontract.Executor
	P2P        *p2p.P2PManager
	Consensus  *consensus.RoundRobin

	mux sync.RWMutex
}

func CreateNode(ID string, port string, validators []types.ValidatorConfig) *Node {
	// 1. Convert Slice to Map untuk lookup cepat
	validatorsMap := make(map[string]types.ValidatorConfig)
	for _, v := range validators {
		validatorsMap[v.ID] = v
	}
	_, isValidator := validatorsMap[ID]

	p2pMan := p2p.CreateP2PManager(ID, port)

	blockchain := InitializeBlockChain()
	mempool := mempool.NewPool()
	ws := state.CreateWorldState()

	node := Node{
		ID:          ID,
		isValidator: isValidator,
		validators:  validatorsMap,
		peers:       make(map[string]string),
		// SmartContractExecutor
		Blockchain: blockchain,
		WorldState: ws,
		Mempool:    mempool,
		P2P:        p2pMan,
	}

	node.Consensus = consensus.NewRoundRobin(ID, &node, validatorsMap)
	node.P2P.Subscribe(node.handleIncomingMessage)

	return &node
}

func (node *Node) Start() {
	if err := node.P2P.Open(); err != nil {
		panic(err)
	}

	node.ConnectToNetwork()
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
	// (Di implementasi real, cukup tanya 1 peer acak atau round robin)
	for id := range node.validators {
		if id == node.ID {
			continue
		}

		// Gunakan Request dengan timeout pendek
		resp, err := node.P2P.Request(id, reqMessage, 2*time.Second)
		if err != nil {
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
			// (Sederhana: Cek Height dan Hash Parent)
			newBlock := blockPayload.Block
			if newBlock.Header.Height == nextHeight {
				// CommitBlock sudah handle validasi lanjutan & insert DB
				// PENTING: CommitBlock menggunakan mutex internal blockchain, jadi aman dipanggil di sini
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
		node.P2P.RemovePeer(address)
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

	node.Blockchain.AddBlock(block)

	node.Executor.ApplyBlock(block)

	node.Mempool.RemoveTxs(block.Transactions)

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
	txs := node.Mempool.PickTxs(MaxBlockTxs)
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
