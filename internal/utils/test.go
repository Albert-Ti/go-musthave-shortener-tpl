package utils

import "github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"

func GenerateMockUUID() func() {
	original := service.GenerateUUID
	counter := 0

	service.GenerateUUID = func() string {
		counter++
		return "key_" + string(rune('0'+counter))
	}

	return func() {
		service.GenerateUUID = original
	}
}
