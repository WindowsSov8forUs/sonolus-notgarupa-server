package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func Generate() (string, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}

	hexID := hex.EncodeToString(bytes[:])
	return fmt.Sprintf("%s-%s-%s", hexID[:8], hexID[8:12], hexID[12:16]), nil
}
