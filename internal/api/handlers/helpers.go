package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// generateID generates a unique ID for new entities.
// In production, this should use a proper UUID library.
func generateID() string {
	// Generate 16 random bytes (128 bits)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return time.Now().Format("20060102150405") + hex.EncodeToString(bytes[:4])
	}
	return hex.EncodeToString(bytes)
}
