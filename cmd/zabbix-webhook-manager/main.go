package main

import (
	"log"
	"net/http"

	"zabbix-webhook-manager/internal/config"
	"zabbix-webhook-manager/internal/httpserver"
	"zabbix-webhook-manager/internal/webhook"
)

func main() {
	cfg, cfgPath, cfgLoaded, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", cfgPath, err)
	}
	if !cfgLoaded {
		log.Printf("Config file not found at %s, using defaults and environment overrides", cfgPath)
	} else {
		log.Printf("Loaded config from %s", cfgPath)
	}

	log.Printf(
		"Config: port=%s endpoints=%d max_body_bytes=%d log_headers=%t log_body=%t https=%t",
		cfg.Port,
		len(cfg.Endpoints),
		cfg.MaxBodyBytes,
		cfg.LogHeaders,
		cfg.LogBody,
		cfg.HTTPSEnabled,
	)

	mux := http.NewServeMux()
	for _, endpoint := range cfg.Endpoints {
		handler := webhook.Handler{
			EndpointName: endpoint.Name,
			EndpointPath: endpoint.Path,
			LogHeaders:   cfg.LogHeaders,
			LogBody:      cfg.LogBody,
			MaxBodyBytes: cfg.MaxBodyBytes,
		}
		log.Printf("Registering webhook endpoint: name=%q path=%q", endpoint.Name, endpoint.Path)
		mux.Handle(endpoint.Path, handler)
	}

	srv := httpserver.Server{
		Addr:        ":" + cfg.Port,
		Handler:     mux,
		EnableHTTPS: cfg.HTTPSEnabled,
		TLSCertFile: cfg.TLSCertFile,
		TLSKeyFile:  cfg.TLSKeyFile,
	}

	log.Fatal(srv.Start())
}
