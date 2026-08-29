package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	shortCodeLength   = 8
	shortCodeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func generateShortCode() (string, error) {
	result := make([]byte, shortCodeLength)

	for i := range result {
		n, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(shortCodeAlphabet))),
		)
		if err != nil {
			return "", fmt.Errorf("generate short code: %w", err)
		}

		result[i] = shortCodeAlphabet[n.Int64()]
	}

	return string(result), nil
}
