package types

// ValidatorConfig digunakan untuk passing data dari main/cmd ke Node
type ValidatorConfig struct {
	ID      string
	Secret  string
	Address string // IP:Port (ex: "192.168.1.5:9000")
}
