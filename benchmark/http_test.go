package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkRoute(b *testing.B) {
	router := setupRouter()
	w := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/ping", nil)
		router.ServeHTTP(w, req)
		assert.Equal(b, 200, w.Code)
	}
}
