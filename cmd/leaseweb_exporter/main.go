package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Nmishin/leaseweb_exporter/internal/client"
	"github.com/Nmishin/leaseweb_exporter/internal/collector"
	"github.com/Nmishin/leaseweb_exporter/internal/config"
)

func main() {
	cfg := config.DefaultRootConfig()
	cfg.FromEnvironment()

	if cfg.ApiKey == "" {
		log.Fatal("LW_EXPORTER_API_KEY environment variable is required")
	}
	client.Init(cfg.ApiKey)

	http.HandleFunc("/metrics", collector.MetricsHandler)
	http.HandleFunc("/targets", collector.TargetsHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	listenAddr := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	log.Printf("Starting exporter on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}
