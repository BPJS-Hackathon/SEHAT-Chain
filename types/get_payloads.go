// Menyimpan payload berdasarkan request / GET dari frontend
package types

import (
	"encoding/json"
)

// Wrapper
type ResponsePayload struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

// Payload untuk membalas request melihat apakah ada rujukan aktif
// untuk faskes 2
type RujukanResponse struct {
	ID         string `json:"id"`
	PesertaID  string `json:"peserta_id"`
	Status     string `json:"status"`
	ExpiryDate int64  `json:"expiry_date"` // Unix timestamp
}

const (
	RujukanStatusActive  = "ACTIVE"  // Aktif
	RujukanStatusExpired = "EXPIRED" // Lewat tanggal exp
	RujukanStatusUsed    = "USED"    // Telah diclaim faskes
)
