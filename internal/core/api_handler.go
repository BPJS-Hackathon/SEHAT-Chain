package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bpjs-hackathon/sehat-chain/types"
	"github.com/google/uuid"
)

func (node *Node) handleFK1RekamMedisPost(w http.ResponseWriter, r *http.Request) {
	var reqData FK1RMSubmitRequest

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	toSum := fmt.Sprintf("%s:%s:%s:%s:%s", reqData.RekamMedisID, reqData.PesertaNIK, reqData.UserID, reqData.DiagnosisCode, reqData.Outcome)
	toSumByte := sha256.Sum256([]byte(toSum))
	rmHash := hex.EncodeToString(toSumByte[:])

	timeStamp := time.Now().Unix()

	// Buat rekaman kunjungan
	visitPayload := types.TxVisit{
		RekamMedisID:   reqData.RekamMedisID,
		RekamMedisHash: rmHash,
	}
	visitJson, _ := json.Marshal(visitPayload)
	tx := types.Transaction{
		ID:        uuid.NewString(),
		Type:      types.TxTypeRecordVisit,
		Timestamp: timeStamp,
		SenderID:  node.ID,
		Payload:   visitJson,
	}
	tx.Signature = node.SignData([]byte(tx.Hash()))
	node.submitTransactionToNetwork(tx)

	// Tidak perlu buat rujukan jika pasien sembuh di FK1
	if reqData.Outcome == "SEMBUH" {
		return
	}

	rujukanID := uuid.NewString()
	txPayload := types.TxRujukan{
		RujukanID:       rujukanID,
		PesertaID:       reqData.PesertaNIK,
		FaskesPembuatID: reqData.FaskesPembuatID,
		FaskesTujuanID:  reqData.FaskesTujuanID,
		RekamMedisID:    reqData.RekamMedisID,
		RekamMedisHash:  rmHash,
		DiagnosisCode:   reqData.RekamMedisID,
	}
	txJson, _ := json.Marshal(txPayload)

	tx = types.Transaction{
		ID:        uuid.NewString(),
		Type:      types.TxTypeCreateRujukan,
		Timestamp: timeStamp,
		SenderID:  node.ID,
		Payload:   txJson,
	}
	tx.Signature = node.SignData([]byte(tx.Hash()))

	node.submitTransactionToNetwork(tx)

	response := FK1RMSubmitResponse{
		RujukanID: "TODO",
	}
	responseJson, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}

func (node *Node) handleFK2RekamMedisPost(w http.ResponseWriter, r *http.Request) {
	var reqData FK2SubmitRequest

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		fmt.Printf("FK2 posted rekam medis but cannot decode")
	}

	timeStamp := time.Now().Unix()

	toSum := fmt.Sprintf("%s:%s:%s:%s:%s", reqData.RekamMedisID, reqData.PesertaNIK, reqData.UserID, reqData.DiagnosisCode, reqData.Outcome)
	toSumByte := sha256.Sum256([]byte(toSum))
	rmHash := hex.EncodeToString(toSumByte[:])

	// Buat rekaman kunjungan
	visitPayload := types.TxVisit{
		RekamMedisID:   reqData.RekamMedisID,
		RekamMedisHash: rmHash,
	}
	visitJson, _ := json.Marshal(visitPayload)
	tx := types.Transaction{
		ID:        uuid.NewString(),
		Type:      types.TxTypeRecordVisit,
		Timestamp: timeStamp,
		SenderID:  node.ID,
		Payload:   visitJson,
	}
	tx.Signature = node.SignData([]byte(tx.Hash()))
	node.submitTransactionToNetwork(tx)

	// FK2 otomatis membuat rekaman claim
	txPayload := types.TxSubmitClaim{
		ClaimID:        reqData.ClaimID,
		RujukanID:      reqData.RujukanID,
		RekamMedisID:   reqData.RekamMedisID,
		RekamMedisHash: rmHash,
		DiagnosisCode:  reqData.DiagnosisCode,
		Amount:         reqData.Amount,
	}

	txJson, _ := json.Marshal(txPayload)

	tx = types.Transaction{
		ID:        uuid.NewString(),
		Type:      types.TxTypeSubmitClaim,
		Timestamp: time.Now().Unix(),
		SenderID:  node.ID,
		Payload:   txJson,
	}
	tx.Signature = node.SignData([]byte(tx.Hash()))

	node.submitTransactionToNetwork(tx)

	// Send empty header if OK
	w.WriteHeader(http.StatusNoContent)
}

func (node *Node) handleClaimExecute(w http.ResponseWriter, r *http.Request) {
	var reqData ExecuteClaim

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Buat tx eksekusi claim
	executePayload := types.TxExecuteClaim{
		ClaimID: reqData.ClaimID,
		Status:  reqData.Status,
	}
	executeJson, _ := json.Marshal(executePayload)

	tx := types.Transaction{
		ID:        uuid.NewString(),
		Type:      types.TxTypeExecuteClaim,
		Timestamp: time.Now().Unix(),
		SenderID:  node.ID,
		Payload:   executeJson,
	}
	tx.Signature = node.SignData([]byte(tx.Hash()))

	node.submitTransactionToNetwork(tx)

	w.WriteHeader(http.StatusNoContent)
}

func (node *Node) handleBlockTotalReq(w http.ResponseWriter, _ *http.Request) {
	type BlockCount struct {
		Count uint64 `json:"count"`
	}

	payload := BlockCount{
		Count: node.Blockchain.GetLatestHeight() - 1, // Ignore genesis
	}
	payloadJson, _ := json.Marshal(payload)

	w.WriteHeader(http.StatusOK)
	w.Write(payloadJson)
}

func (node *Node) handleAPIBlockRequest(w http.ResponseWriter, r *http.Request) {
	type BlockResp struct {
		types.Block
	}

	heightStr := r.PathValue("height")

	height, err := strconv.Atoi(heightStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if height < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	block, err := node.Blockchain.GetBlock(uint64(height))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload := BlockResp{
		Block: block,
	}
	payloadJson, _ := json.Marshal(payload)

	w.WriteHeader(http.StatusOK)
	w.Write(payloadJson)
}
