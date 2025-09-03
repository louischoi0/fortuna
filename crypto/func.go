package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

