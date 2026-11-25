package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func hashRequest(request any) (string, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}
