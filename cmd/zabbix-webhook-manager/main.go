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
		"Config: port=%s path=%s max_body_bytes=%d log_headers=%t log_body=%t",
		cfg.Port,
		cfg.WebhookPath,
		cfg.MaxBodyBytes,
		cfg.LogHeaders,
		cfg.LogBody,
	)

	handler := webhook.Handler{
		LogHeaders:   cfg.LogHeaders,
		LogBody:      cfg.LogBody,
		MaxBodyBytes: cfg.MaxBodyBytes,
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.WebhookPath, handler)

	srv := httpserver.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	log.Fatal(srv.Start())
}
