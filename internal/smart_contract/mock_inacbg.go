package smartcontract

import (
	"fmt"
	"strings"
)

// MockInaCBGValidator mensimulasikan pengecekan ke server INA-CBG
type MockInaCBGValidator struct{}

func NewInaCBGValidator() *MockInaCBGValidator {
	return &MockInaCBGValidator{}
}

// VerifyClaim mengecek apakah diagnosis code valid dan eligible untuk diklaim
func (v *MockInaCBGValidator) VerifyClaim(diagnosisCode string) (bool, uint64, error) {
	// Simulasi: Diagnosis code harus diawali huruf tertentu (misal 'A')
	// dan amount tidak boleh 0
	if len(diagnosisCode) == 0 {
		return false, 0, fmt.Errorf("diagnosis code is empty")
	}

	code := strings.ToUpper(diagnosisCode)

	if !strings.HasPrefix(strings.ToUpper(code), "A") {
		return false, 0, fmt.Errorf("invalid diagnosis code prefix")
	}

	exists, amount := v.SearchCode(code)
	if !exists {
		return false, 0, fmt.Errorf("diagnosis code unavailable")
	}

	return true, amount, nil
}

func (v *MockInaCBGValidator) SearchCode(diagnosisCode string) (bool, uint64) {
	mockCodes := map[string]uint64{
		"A00": 500000,
		"A01": 750000,
		"A02": 600000,
		"A03": 550000,
		"A04": 450000,
		"A05": 350000,
		"A06": 800000,
		"A07": 700000,
		"A08": 400000,
		"A09": 300000,
	}

	amount, exists := mockCodes[diagnosisCode]
	return exists, amount
}
