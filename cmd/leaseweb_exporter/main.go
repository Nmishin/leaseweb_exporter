package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Nmishin/leaseweb_exporter/internal/client"
)

func main() {
	apiKey := os.Getenv("LEASEWEB_API_KEY")
	if apiKey == "" {
		log.Fatal("LEASEWEB_API_KEY is not set")
	}
	client.Init(apiKey)

	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("Listening on :9112")
	log.Fatal(http.ListenAndServe(":9112", nil))
}
