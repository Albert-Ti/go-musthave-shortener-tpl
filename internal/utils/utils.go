package utils

import (
	"crypto/rand"
)

var GenerateUUID = func() string {

	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 8

	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	for i := range bytes {
		bytes[i] = alphabet[int(bytes[i])%len(alphabet)]
	}

	return string(bytes)
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
