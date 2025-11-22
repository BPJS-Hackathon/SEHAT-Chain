package state

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"github.com/bpjs-hackathon/sehat-chain/types"
)

type WorldState struct {
	VisitRecord map[string]types.TxVisit
	Rujukans    map[string]types.RujukanAsset
	Claims      map[string]types.ClaimAsset
	mux         sync.RWMutex
}

func CreateWorldState() *WorldState {
	return &WorldState{
		Rujukans: make(map[string]types.RujukanAsset),
		Claims:   make(map[string]types.ClaimAsset),
	}
}

func (ws *WorldState) AddVisit(visit types.TxVisit) {
	ws.mux.Lock()
	defer ws.mux.Unlock()

	ws.VisitRecord[visit.RekamMedisID] = visit
}

func (ws *WorldState) AddClaim(claim types.ClaimAsset) {
	ws.mux.Lock()
	defer ws.mux.Unlock()

	ws.Claims[claim.ClaimID] = claim
}

func (ws *WorldState) AddRujukan(rujukan types.RujukanAsset) {
	ws.mux.Lock()
	defer ws.mux.Unlock()

	ws.Rujukans[rujukan.ID] = rujukan
}

func (ws *WorldState) GetVisit(rekamMedisID string) (types.TxVisit, bool) {
	ws.mux.RLock()
	defer ws.mux.RUnlock()

	visit, exists := ws.VisitRecord[rekamMedisID]
	return visit, exists
}

func (ws *WorldState) GetClaim(claimID string) (types.ClaimAsset, bool) {
	ws.mux.RLock()
	defer ws.mux.RUnlock()

	claim, exists := ws.Claims[claimID]
	return claim, exists
}

func (ws *WorldState) GetRujukan(rujukanID string) (types.RujukanAsset, bool) {
	ws.mux.RLock()
	defer ws.mux.RUnlock()

	rujukan, exists := ws.Rujukans[rujukanID]
	return rujukan, exists
}

func (ws *WorldState) CalculateHash() string {
	ws.mux.Lock()
	defer ws.mux.Unlock()

	// Sort asset agar hash deterministic
	var rujukansArr []string
	var claimsArr []string

	for k := range ws.Rujukans {
		rujukansArr = append(rujukansArr, k)
	}

	for k := range ws.Claims {
		claimsArr = append(claimsArr, k)
	}

	sort.Strings(rujukansArr)
	sort.Strings(claimsArr)

	var combinedData string
	for _, k := range rujukansArr {
		rujukan := ws.Rujukans[k]
		combinedData += fmt.Sprintf("%s:%s:%s|", rujukan.ID, rujukan.RekamMedisHash, rujukan.Status)
	}
	for _, k := range claimsArr {
		claim := ws.Claims[k]
		combinedData += fmt.Sprintf("%s:%s:%s|", claim.ClaimID, claim.RekamMedisHash, claim.Status)
	}

	hash := sha256.Sum256([]byte(combinedData))
	return hex.EncodeToString(hash[:])
}
