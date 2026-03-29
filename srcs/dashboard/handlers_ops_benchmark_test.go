package dashboard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onehumancorp/mono/srcs/orchestration"
)

func BenchmarkScaleStream(b *testing.B) {
	hub := orchestration.NewHub()
	s := &Server{
		hub: hub,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/scale/stream", nil)
		w := httptest.NewRecorder()
		s.handleScaleStream(w, req)
	}
}
