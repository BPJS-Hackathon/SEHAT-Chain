package core

import "github.com/bpjs-hackathon/sehat-chain/types"

type RekamMedis struct {
	RekamMedisID  string `json:"id"`
	PesertaNIK    string `json:"peserta_nik"`
	UserID        string `json:"user_id"`
	DiagnosisCode string `json:"diagnosis_code"`
	Note          string `json:"note"`
	JenisRawat    string `json:"jenis_rawat"`
	AdmissionDate int64  `json:"admission_date"`
	DischargeDate int64  `json:"discharge_date"`
	Outcome       string `json:"outcome"`
}

// /// /// /// /// /// /// ///  //
// Faskes 1 Kirim ke blockchain //
// /// /// /// /// /// /// /// //
type Rujukan struct {
	FaskesPembuatID string `json:"faskes_pembuat"`
	FaskesTujuanID  string `json:"faskes_tujuan"`
}

type FK1RMSubmitRequest struct {
	RekamMedis
	// Rujukan // Hardcoded
}

// Return ID rujukan yang dibuat (untuk dibawa ke pasien)
type FK1RMSubmitResponse struct {
	RujukanID string `json:"rujukan_id"`
}

// /// /// /// /// /// /// ///  //
// Faskes 2 Kirim ke blockchain //
// /// /// /// /// /// /// /// //
type Claim struct {
	ClaimID string `json:"claim_id"`
	Amount  uint64 `json:"amount"`
}

type FK2SubmitRequest struct {
	RekamMedis
	RujukanID string `json:"rujukan_id"`
	Claim
}

// /// /// /// /// /// /// /// //
// Faskes 1/2 Request Rujukan  //
// /// /// /// /// /// /// /// //
type GetRujukanInfo struct {
	types.RujukanAsset
}

// /// /// /// /// /// ///
// Admin Execute Claim  //
// /// /// /// /// /// ///
type ExecuteClaim struct {
	Claim
	Status string `json:"status"` // PAID or REJECTED
}

// /// /// /// /// /// /// ///  //
// Verify Klaim Status By All  //
// /// /// /// /// /// /// /// //
type GetClaimInfo struct {
	types.ClaimAsset
}
