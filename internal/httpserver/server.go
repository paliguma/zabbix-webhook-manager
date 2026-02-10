package httpserver

import (
	"log"
	"net/http"
)

type Server struct {
	Addr    string
	Handler http.Handler
}

func (s Server) Start() error {
	log.Printf("Listening on %s", s.Addr)
	return http.ListenAndServe(s.Addr, s.Handler)
}
