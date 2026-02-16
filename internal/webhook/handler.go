package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
	EndpointName string
	EndpointPath string
	LogHeaders   bool
	LogBody      bool
	MaxBodyBytes int64
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.MaxBodyBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, h.MaxBodyBytes)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	endpoint := h.endpointLabel(r.URL.Path)

	log.Printf("Webhook received: endpoint=%q method=%s path=%s", endpoint, r.Method, r.URL.Path)
	log.Printf("Remote: %s", r.RemoteAddr)
	if h.LogHeaders {
		log.Printf("Headers:\n%s", formatHeaders(r.Header))
	}
	if h.LogBody {
		log.Printf("Body:\n%s", string(body))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":   "ok",
		"endpoint": endpoint,
	})
}

func formatHeaders(h http.Header) string {
	var b strings.Builder
	for k, v := range h {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(strings.Join(v, ", "))
		b.WriteString("\n")
	}
	return b.String()
}

func (h Handler) endpointLabel(requestPath string) string {
	if h.EndpointName != "" {
		return h.EndpointName
	}
	if h.EndpointPath != "" {
		return h.EndpointPath
	}
	return requestPath
}
