package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign creates a signature from the payload using the endpoint's secret
func Sign(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// Verify checks if a received signature is valid
func Verify(secret string, payload []byte, receivedSig string) bool {
	expected := Sign(secret, payload)
	return hmac.Equal([]byte(expected), []byte(receivedSig))
}
