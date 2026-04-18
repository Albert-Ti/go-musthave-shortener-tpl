package handler_test

import (
	"net/http"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func TestCreateShortUrlJSON(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		urlService *service.URLService
		want       http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.CreateShortUrlJSON(tt.urlService)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("CreateShortUrlJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
