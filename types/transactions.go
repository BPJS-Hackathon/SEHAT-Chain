// Wrapping transaction message untuk dikirim ke antar node blockchain
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// TransactionType
const (
	TxTypeMintReferral = "MINT_REFERRAL" // Menciptakan token rujukan
	TxTypeRecordVisit  = "RECORD_VISIT"  // Mencatat kunjungan TANPA rujukan
	TxTypeSubmitClaim  = "SUBMIT_CLAIM"  // Submit klaim oleh pengunjung
	TxTypeExecuteClaim = "EXECUTE_CLAIM" // Persetujuan final bahwa BPJS telah memvalidasi dan akan membayar
)

// Wrapping transaction yang disebar antar node
type Transaction struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
	SenderID  string          `json:"sender_id"`
	Signature string          `json:"secret"` // Mock signature
	Payload   json.RawMessage `json:"payload"`
}

func (tx *Transaction) Hash() string {
	data := fmt.Sprintf("%s%d%s%s",
		tx.Type,
		tx.Timestamp,
		tx.SenderID,
		string(tx.Payload),
	)

	hashBytes := sha256.Sum256([]byte(data))
	hash := hex.EncodeToString(hashBytes[:])
	return hash
}
