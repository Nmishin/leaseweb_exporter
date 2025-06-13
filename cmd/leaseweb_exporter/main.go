package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Nmishin/leaseweb_exporter/internal/client"
	"github.com/Nmishin/leaseweb_exporter/internal/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	apiKey := os.Getenv("LEASEWEB_API_KEY")
	if apiKey == "" {
		log.Fatal("LEASEWEB_API_KEY is not set")
	}
	client.Init(apiKey)

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector.NewDedicatedServerCollector())

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Println("Listening on :9112")
	log.Fatal(http.ListenAndServe(":9112", nil))
}

