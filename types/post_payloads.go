// Menyimpan payload yang dikirim oleh frontend
package types

// Payload ketika faskes 1 upload rekam medis tanpa membuat rujuk
type TxVisit struct {
	RekamMedisID   string `json:"rekam_medis_id"`
	RekamMedisHash string `json:"rekam_medis_hash"`
}

// Payload pembuatan referral (rujukan)
// digunakan oleh faskes 1 saat upload rekam medis dengan outcome rujukan
type TxRujukan struct {
	RujukanID       string `json:"rujukan_id"`
	PesertaID       string `json:"patient_id"`
	RekamMedisID    string `json:"rekam_medis_id"`
	RekamMedisHash  string `json:"rekam_medis_hash"`
	FaskesPembuatID string `json:"origin_faskes_id"`
	FaskesTujuanID  string `json:"target_faskes_id"`
	DiagnosisCode   string `json:"diagnosis_code"`
}

// Payload untuk submit claim
// digunakan oleh rumah sakit terujuk saat upload rekam medis final.
type TxSubmitClaim struct {
	ClaimID        string `json:"claim_id"`
	RujukanID      string `json:"rujukan_id"`
	RekamMedisID   string `json:"rekam_medis_id"`
	RekamMedisHash string `json:"rekam_medis_hash"`
	DiagnosisCode  string `json:"diagnosis_final"`
}

// Payload persetujuan pembayaran
// Scenario: submit claim di cek pada ina-cbg dengan smart contract
// List claim dan pembayaran pending di verifikasi admin
// server mengirim bukti pembayaran disimpan di blockchain
// So digunakan server BPJS untuk siap wiring uang ke faskes
type TxExecuteClaim struct {
	ClaimID string `json:"claim_id"`
	AdminID string `json:"admin_id"`
	Status  string `json:"status"`
}

const (
	ClaimStatusSubmitted = "SUBMITTED" // Claim disubmit client tapi belum diverifikasi blockchain
	ClaimStatusPending   = "PENDING"   // Claim sudah diverifikasi blockchain
	ClaimStatusPaid      = "PAID"      // TxExecuteClaim sudah dijalankan admin
	ClaimStatusRejected  = "REJECTED"  // Claim direject oleh blockchain atau saat TxExecuteClaim
)
