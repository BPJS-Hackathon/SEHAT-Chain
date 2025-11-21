package utils

import "fmt"

// Mock keypair
type CryptoCred struct {
	secret string
}

func NewCred(secret string) *CryptoCred {
	return &CryptoCred{
		secret: secret,
	}
}

func (cc *CryptoCred) GetSecret() string {
	return cc.secret
}

func (cc *CryptoCred) Sign(data []byte) string {
	return cc.secret
}

func (cc *CryptoCred) Validate(signature string, data []byte) error {
	if signature != cc.secret {
		return fmt.Errorf("wrong signature, has (%s) but calculated (%s)", signature, cc.secret)
	}

	return nil
}
