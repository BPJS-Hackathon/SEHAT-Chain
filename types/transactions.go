package types

import "encoding/json"

// TransactionType
const (
	TxMintReferral   = "MINT_REFERRAL" // Menciptakan token rujukan
	TxRecordVisit    = "RECORD_VISIT"  // Mencatat kunjungan TANPA rujukan
	TxSubmitClaim    = "SUBMIT_CLAIM"  // Submit klaim oleh pengunjung
	TxExecutePayment = "EXEC_PAYMENT"  // Persetujuan final bahwa BPJS telah memvalidasi dan akan membayar
)

type Transaction struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
	// Signature string `json:"signature"` // Optional
}
