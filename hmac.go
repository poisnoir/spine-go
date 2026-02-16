package spine

import (
	"crypto/hmac"
	"crypto/sha256"
)

func generateHmac(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func verifyHmac(key []byte, data []byte, expected []byte) bool {
	actual := generateHmac(key, data)
	return hmac.Equal(actual, expected)
}
