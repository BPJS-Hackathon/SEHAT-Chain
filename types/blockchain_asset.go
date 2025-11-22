// Berisi data yang tersimpan di blockchain
package types

type RujukanAsset struct {
	ID        string `json:"id"`
	PesertaID string `json:"peserta_id"`

	FaskesPembuatID string `json:"faskes_pembuat_id"`
	FaskesTujuanID  string `json:"faskes_tujuan_id"`

	RekamMedisID   string `json:"rekam_medis_id"`
	RekamMedisHash string `json:"rekam_medis_hash"`

	Status     string `json:"status"`
	IssueDate  int64  `json:"issue_date"`  // Unix timestamp
	ExpiryDate int64  `json:"expiry_date"` // Unix timestamp
}

type ClaimAsset struct {
	ClaimID string `json:"claim_id"`

	RujukanID string `json:"rujukan_id"`
	FaskesID  string `json:"faskes_id"`

	RekamMedisID   string `json:"rekam_medis_id"`
	RekamMedisHash string `json:"rekam_medis_hash"`

	DiagnosisCode string `json:"diagnosis_code"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	Timestamp     int64  `json:"timestamp"` // Kapan disubmit
}
