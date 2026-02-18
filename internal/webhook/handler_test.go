package webhook

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_AllowsRequestFromAllowedIP(t *testing.T) {
	handler := Handler{
		EndpointName:   "test-endpoint",
		EndpointPath:   "/webhook/test",
		AllowedSources: []string{"192.168.1.10", "10.0.0.0/24"},
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook/test", strings.NewReader(`{"ping":"pong"}`))
	req.RemoteAddr = "192.168.1.10:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_RejectsRequestFromNotAllowedIP(t *testing.T) {
	handler := Handler{
		EndpointName:   "test-endpoint",
		EndpointPath:   "/webhook/test",
		AllowedSources: []string{"192.168.1.10", "10.0.0.0/24"},
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook/test", strings.NewReader(`{"ping":"pong"}`))
	req.RemoteAddr = "8.8.8.8:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}
