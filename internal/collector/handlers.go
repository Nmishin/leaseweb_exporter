package collector

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type TargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "Target parameter is required", http.StatusBadRequest)
		return
	}

	coll, err := GetCollector(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(coll)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{Timeout: 10 * time.Second})
	h.ServeHTTP(w, r)
}

func TargetsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	d, err := GetCollector("__discovery_service__")
	if err != nil {
		http.Error(w, "Internal error", 500)
		return
	}

	ids, err := d.GetAllServerIDs(ctx)
	if err != nil {
		log.Printf("Discovery error: %v", err)
		http.Error(w, "Discovery failed", http.StatusBadGateway)
		return
	}

	tgs := []TargetGroup{{
		Targets: ids,
		Labels:  map[string]string{"provider": "leaseweb"},
	}}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tgs)
}
