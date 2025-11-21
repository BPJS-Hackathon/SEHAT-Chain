package types

type QuorumCertificate struct {
	HeaderHash string            `json:"header_hash"`
	Signatures map[string]string `json:"signatures"` // map signature berdasarkan id
}
