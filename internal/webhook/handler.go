package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type Handler struct {
	EndpointName   string
	EndpointPath   string
	AllowedSources []string
	LogHeaders     bool
	LogBody        bool
	MaxBodyBytes   int64
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	endpoint := h.endpointLabel(r.URL.Path)
	clientIP, err := parseRemoteIP(r.RemoteAddr)
	if err != nil {
		log.Printf(
			"Webhook rejected: endpoint=%q path=%s remote=%q reason=invalid remote address",
			endpoint,
			r.URL.Path,
			r.RemoteAddr,
		)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !h.isSourceAllowed(clientIP) {
		log.Printf(
			"Webhook rejected: endpoint=%q path=%s remote=%q client_ip=%s reason=source not allowed",
			endpoint,
			r.URL.Path,
			r.RemoteAddr,
			clientIP.String(),
		)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

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

	log.Printf("Webhook received: endpoint=%q method=%s path=%s", endpoint, r.Method, r.URL.Path)
	log.Printf("Remote: %s (client_ip=%s)", r.RemoteAddr, clientIP.String())
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

func (h Handler) isSourceAllowed(clientIP net.IP) bool {
	if len(h.AllowedSources) == 0 {
		return true
	}

	for _, source := range h.AllowedSources {
		if strings.Contains(source, "/") {
			_, network, err := net.ParseCIDR(source)
			if err != nil {
				continue
			}
			if network.Contains(clientIP) {
				return true
			}
			continue
		}

		allowedIP := net.ParseIP(source)
		if allowedIP == nil {
			continue
		}
		if allowedIP.Equal(clientIP) {
			return true
		}
	}

	return false
}

func parseRemoteIP(remoteAddr string) (net.IP, error) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// Fallback for non host:port values.
		host = remoteAddr
	}

	// Remove IPv6 zone if present (for example fe80::1%eth0).
	host = strings.Split(host, "%")[0]

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, net.InvalidAddrError("invalid remote address")
	}
	return ip, nil
}
