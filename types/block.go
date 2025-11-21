// Berisi struktur block
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type Block struct {
	Header       BlockHeader       `json:"header"`
	Transactions []Transaction     `json:"transactions"`
	QC           QuorumCertificate `json:"qc"`
}

type BlockHeader struct {
	Height     uint64 `json:"height"`
	Timestamp  int64  `json:"timestamp"`
	PrevHash   string `json:"prev_hash"`
	StateRoot  string `json:"state_root"`
	TxRoot     string `json:"tx_root"`
	ProposerID string `json:"proposer"`
}

func (b *Block) HeaderHash() string {
	data := fmt.Sprintf("%d%d%s%s%s%s",
		b.Header.Height,
		b.Header.Timestamp,
		b.Header.PrevHash,
		b.Header.StateRoot,
		b.Header.TxRoot,
		b.Header.ProposerID,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func CalculateTxRoot(txs []Transaction) string {
	if len(txs) == 0 {
		return strings.Repeat("0", 64)
	}

	h := sha256.New()
	for _, tx := range txs {
		txHash := tx.Hash()
		txHashBytes, _ := hex.DecodeString(txHash)

		h.Write(txHashBytes)
	}

	return hex.EncodeToString(h.Sum(nil))
}
