package consensus

type QuorumCertificate struct {
	HeaderHash []byte            `json:"header_hash"`
	Signatures map[string]string `json:"signatures"` // map signature berdasarkan id
}
