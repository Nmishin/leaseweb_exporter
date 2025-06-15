package main

import (
	"net/http"

	"github.com/Nmishin/leaseweb_exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "target parameter is required", http.StatusBadRequest)
		return
	}

	collector, err := collector.GetCollector(target)
	if err != nil {
		http.Error(w, "could not get collector: "+err.Error(), http.StatusInternalServerError)
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func healthHandler(rsp http.ResponseWriter, req *http.Request) {
	// just return a simple 200 for now
}
