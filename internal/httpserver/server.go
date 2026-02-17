package httpserver

import (
	"log"
	"net/http"
)

type Server struct {
	Addr        string
	Handler     http.Handler
	EnableHTTPS bool
	TLSCertFile string
	TLSKeyFile  string
}

func (s Server) Start() error {
	if s.EnableHTTPS {
		log.Printf("Listening with HTTPS on %s", s.Addr)
		return http.ListenAndServeTLS(s.Addr, s.TLSCertFile, s.TLSKeyFile, s.Handler)
	}

	log.Printf("Listening with HTTP on %s", s.Addr)
	return http.ListenAndServe(s.Addr, s.Handler)
}
