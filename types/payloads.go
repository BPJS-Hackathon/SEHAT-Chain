package types

// Payload pembuatan referral (rujukan)
// digunakan oleh faskes 1
type CreateReferralRequest struct {
	RequestID      string `json:"request_id"`
	PatientID      string `json:"patient_id"`
	OriginFaskesID string `json:"origin_faskes_id"`
	TargetFaskesID string `json:"target_faskes_id"`
	//	Opsional. Harus dibuatin table daftar diagnosis :?
	// 	TargetPoli string `json:"target_poli"`
	// 	DiagnosisCode string `json:"diagnosis_code"`
}

// Payload untuk submit claim
// digunakan oleh rumah sakit terujuk.
// note: terujuk tidak perlu submit record visit,
// anggapan submit claim = record visit
// submit claim
type SubmitClaimRequest struct {
	ClaimRequestID string `json:"claim_request_id"`
	ReferralID     string `json:"referral_id"` //cukup refer ke rujukan
	FaskesID       string `json:"hospital_id"` // rs/faskes pensubmit klaim
	// Opsional again. Tapi harus kalau mau implement smart contract
	// DiagnosisFinal string `json:"diagnosis_final"` // Diagnosis final pasien
	// Tindakan string `json:"tindakan"` // Opsional. Deskripsi tertulis
	// Outcome string `json:"outcome"` // "SEMBUH", "MENINGGAL"
}

// Payload record visit
// Digunakan faskes 1 jika tanpa rujukan lanjut
type RecordVisitRequest struct {
	VisitID   string `json:"visit_id"`
	PatientID string `json:"patient_id"`
	FaskesID  string `json:"hospital_id"`
	// Diagnosis string `json:"diagnosis"`
	// Outcome string `json:"outcome"` // "SEMBUH", "MENINGGAL", "RUJUKAN"
}

// Payload persetujuan pembayaran
// Scenario: submit claim di cek pada ina-cbg dengan smart contract
// List claim dan pembayaran pending di verifikasi admin
// server mengirim bukti pembayaran disimpan di blockchain
// So digunakan server BPJS untuk siap wiring uang ke faskes
type ExecutePaymentRequest struct {
	ClaimID string `json:"claim_id"`
	// Digunakan hanya jika akan menambahkan banyak admin yang bisa akses server
	// AdminID string `json:"admin_id"`
}
