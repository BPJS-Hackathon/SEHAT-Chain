package smartcontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bpjs-hackathon/sehat-chain/internal/state"
	"github.com/bpjs-hackathon/sehat-chain/types"
)

type Executor struct {
	WorldState *state.WorldState
	InaCBG     *MockInaCBGValidator
}

func NewExecutor(ws *state.WorldState) *Executor {
	return &Executor{
		WorldState: ws,
		InaCBG:     &MockInaCBGValidator{},
	}
}

func (e *Executor) ApplyBlock(block types.Block) {
	for _, tx := range block.Transactions {
		e.applyTransaction(tx)
	}
}

func (e *Executor) applyTransaction(tx types.Transaction) {
	switch tx.Type {
	case types.TxTypeRecordVisit:
		e.handleRecordVisit(tx)
	case types.TxTypeCreateRujukan:
		e.handleCreateRujukan(tx)
	case types.TxTypeSubmitClaim:
		e.handleSubmitClaim(tx)
	case types.TxTypeExecuteClaim:
		e.handleExecuteClaim(tx)
	default:
		fmt.Printf("Unknown transaction type: %s\n", tx.Type)
	}
}

// handleRecordVisit: Mencatat kunjungan pasien biasa
func (e *Executor) handleRecordVisit(tx types.Transaction) {
	var payload types.TxVisit
	if err := json.Unmarshal(tx.Payload, &payload); err != nil {
		fmt.Println("Error unmarshal Visit payload:", err)
		return
	}

	e.WorldState.AddVisit(payload)
	fmt.Printf("✅ [SmartContract] Visit Recorded: %s\n", payload.RekamMedisID)
}

// handleCreateRujukan: Membuat asset rujukan baru
func (e *Executor) handleCreateRujukan(tx types.Transaction) {
	var payload types.TxRujukan
	if err := json.Unmarshal(tx.Payload, &payload); err != nil {
		fmt.Println("Error unmarshal Rujukan payload:", err)
		return
	}

	// Logic: Create Asset Rujukan
	asset := types.RujukanAsset{
		ID:              payload.RujukanID,
		PesertaID:       payload.PesertaID,
		FaskesPembuatID: payload.FaskesPembuatID,
		FaskesTujuanID:  payload.FaskesTujuanID,
		RekamMedisID:    payload.RekamMedisID,
		RekamMedisHash:  payload.RekamMedisHash,
		Status:          types.RujukanStatusActive,
		IssueDate:       time.Now().Unix(),
		ExpiryDate:      time.Now().AddDate(0, 3, 0).Unix(), // Berlaku 3 bulan
	}

	e.WorldState.AddRujukan(asset)
	fmt.Printf("✅ [SmartContract] Rujukan Created: %s -> %s\n", payload.FaskesPembuatID, payload.FaskesTujuanID)
}

// handleSubmitClaim: Faskes mengajukan klaim
func (e *Executor) handleSubmitClaim(tx types.Transaction) {
	var payload types.TxSubmitClaim
	if err := json.Unmarshal(tx.Payload, &payload); err != nil {
		fmt.Println("Error unmarshal Claim payload:", err)
		return
	}

	// 1. Cek Validitas Rujukan (Jika ada rujukan ID)
	if payload.RujukanID != "" {
		rujukan, exists := e.WorldState.GetRujukan(payload.RujukanID)
		if !exists || rujukan.Status != types.RujukanStatusActive {
			fmt.Println("❌ Claim Failed: Rujukan not found or expired")
			e.updateClaimStatusSql(payload.ClaimID, types.ClaimStatusFaked)
			return
		}
		// Tandai rujukan sebagai USED
		rujukan.Status = types.RujukanStatusUsed
		e.WorldState.AddRujukan(rujukan) // Update state
	}

	// 3. Validasi aturan medis
	valid, amount, err := e.InaCBG.VerifyClaim(payload.DiagnosisCode)

	status := types.ClaimStatusPending
	if !valid || err != nil {
		status = types.ClaimStatusRejected
		e.updateClaimStatusSql(payload.ClaimID, types.ClaimStatusRejected)
		fmt.Printf("⚠️ Claim Rejected by Engine: %v\n", err)
	}

	// 4. Buat Asset Claim
	claimAsset := types.ClaimAsset{
		ClaimID:        payload.ClaimID,
		RujukanID:      payload.RujukanID,
		FaskesID:       tx.SenderID,
		RekamMedisID:   payload.RekamMedisID,
		RekamMedisHash: payload.RekamMedisHash,
		DiagnosisCode:  payload.DiagnosisCode,
		Amount:         amount,
		Status:         status,
		Timestamp:      tx.Timestamp,
	}

	e.WorldState.AddClaim(claimAsset)
	fmt.Printf("✅ [SmartContract] Claim Submitted: %s (Status: %s)\n", payload.ClaimID, status)
}

// handleExecuteClaim: Admin BPJS menyetujui pembayaran
func (e *Executor) handleExecuteClaim(tx types.Transaction) {
	var payload types.TxExecuteClaim
	if err := json.Unmarshal(tx.Payload, &payload); err != nil {
		fmt.Println("Error unmarshal Execute Claim payload:", err)
		return
	}

	// Ambil data claim existing
	claim, exists := e.WorldState.GetClaim(payload.ClaimID)
	if !exists {
		fmt.Println("❌ Execute Failed: Claim ID not found")
		return
	}

	// Update status berdasarkan keputusan admin (PAID / REJECTED)
	claim.Status = payload.Status

	e.WorldState.AddClaim(claim)
	e.updateClaimStatusSql(claim.ClaimID, claim.Status)
	fmt.Printf("💰 [SmartContract] Claim Executed: %s is now %s\n", claim.ClaimID, claim.Status)
}

func (e *Executor) updateClaimStatusSql(claimID string, status string) error {
	type UpdateStatus struct {
		Status string `json:"status"`
	}

	payload := UpdateStatus{
		Status: status,
	}

	jsonData, _ := json.Marshal(payload)

	client := &http.Client{}
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost:8080/admin/claims/%s/status", claimID), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("gagal hit database")
	}

	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("gagal hit database")
	}

	return nil
}
