package utils

import (
	"crypto/sha256"
	"fmt"
)

func CreateSha256Checksum(input []byte) (hash string) {
	h := sha256.New()
	h.Write(input)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
