package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Nmishin/leaseweb_exporter/internal/client"
	"github.com/Nmishin/leaseweb_exporter/internal/collector"
)

func main() {
	apiKey := os.Getenv("LEASEWEB_API_KEY")
	if apiKey == "" {
		log.Fatal("LEASEWEB_API_KEY is not set")
	}
	client.Init(apiKey)

	http.HandleFunc("/metrics", collector.MetricsHandler)
	http.HandleFunc("/targets", collector.TargetsHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Leaseweb Exporter listening on :9112")
	if err := http.ListenAndServe(":9112", nil); err != nil {
		log.Fatal(err)
	}
}
