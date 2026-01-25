package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"

	"github.com/Pois-Noir/Botzilla/internal/globals"
)

func GenerateHMAC(data []byte, key []byte) [globals.HASH_LENGTH]byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return [globals.HASH_LENGTH]byte(mac.Sum(nil))
}

func VerifyHMAC(data []byte, key []byte, hash [globals.HASH_LENGTH]byte) bool {
	// Generate HMAC for the provided data using the same key
	generatedHMAC := GenerateHMAC(data, key)

	// Use subtle.ConstantTimeCompare to securely compare the two HMACs
	return subtle.ConstantTimeCompare(generatedHMAC[:], hash[:]) == 1
}
