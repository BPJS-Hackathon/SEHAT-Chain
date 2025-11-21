package types

type QuorumCertificate struct {
	HeaderHash string `json:"header_hash"`
	Signatures string `json:"signatures"` // leader hex encoded signature
}
