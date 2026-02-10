package webhook

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
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

	log.Printf("Webhook received: %s %s", r.Method, r.URL.Path)
	log.Printf("Remote: %s", r.RemoteAddr)
	if h.LogHeaders {
		log.Printf("Headers:\n%s", formatHeaders(r.Header))
	}
	if h.LogBody {
		log.Printf("Body:\n%s", string(body))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
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
