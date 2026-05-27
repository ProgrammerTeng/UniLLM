package crypto

// KeyProtector encrypts and decrypts provider API keys at rest.
type KeyProtector struct {
	encryptionKey string
}

func NewKeyProtector(hexKey string) *KeyProtector {
	return &KeyProtector{encryptionKey: hexKey}
}

func (p *KeyProtector) Encrypt(plaintext string) (string, error) {
	if p.encryptionKey == "" {
		return plaintext, nil
	}
	return Encrypt(p.encryptionKey, plaintext)
}

func (p *KeyProtector) Decrypt(stored string) string {
	if p.encryptionKey == "" {
		return stored
	}
	plain, err := Decrypt(p.encryptionKey, stored)
	if err != nil {
		return stored
	}
	return plain
}

func (p *KeyProtector) Enabled() bool {
	return p.encryptionKey != ""
}
