package reqid

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a random 16-char hex request ID.
func New() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
