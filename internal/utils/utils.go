package utils

import (
	"crypto/rand"
	"encoding/base64"
)

var GenerateUUID = func() string {
	key := make([]byte, 9)
	rand.Read(key)
	return base64.RawURLEncoding.EncodeToString(key)
}

func GenerateMockUUID() func() {
	original := GenerateUUID
	counter := 0

	GenerateUUID = func() string {
		counter++
		return "key_" + string(rune('0'+counter))
	}

	return func() {
		GenerateUUID = original
	}
}
