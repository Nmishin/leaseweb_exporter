package collector

import (
	"context"
	"sync"

	"github.com/Nmishin/leaseweb_exporter/internal/client"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	collectors = make(map[string]*DedicatedServerCollector)
	mu         sync.Mutex
)

type DedicatedServerCollector struct {
	target    string
	collected *sync.Cond
	client    *client.Client

	servers  *prometheus.Desc
	location *prometheus.Desc
}

func NewDedicatedServerCollector(target string) *DedicatedServerCollector {
	return &DedicatedServerCollector{
		target:    target,
		client:    &client.LeasewebClient,
		collected: sync.NewCond(&sync.Mutex{}),
		servers: prometheus.NewDesc(
			"leaseweb_dedicated_server_info",
			"Metadata about a Leaseweb dedicated server",
			[]string{"id", "chassis"}, nil,
		),
		location: prometheus.NewDesc(
			"leaseweb_dedicated_server_location",
			"Server location info",
			[]string{"id", "site"}, nil,
		),
	}
}

func GetCollector(target string) (*DedicatedServerCollector, error) {
	mu.Lock()
	collector, ok := collectors[target]
	if !ok {
		collector = NewDedicatedServerCollector(target)
		collectors[target] = collector
	}
	mu.Unlock()

	// Lock to prevent concurrent scrapes for the same target
	collector.collected.L.Lock()
	defer collector.collected.L.Unlock()

	return collector, nil
}

func (c *DedicatedServerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.servers
	ch <- c.location
}

func (c *DedicatedServerCollector) Collect(ch chan<- prometheus.Metric) {
	resp, _, err := c.client.DedicatedserverAPI.
		GetServer(context.Background(), c.target).
		Execute()
	if err != nil {
		return
	}

	id := deref(resp.Id)
	model := deref(resp.Specs.Chassis)
	site := deref(resp.Location.Site)

	ch <- prometheus.MustNewConstMetric(c.servers, prometheus.GaugeValue, 1, id, model)
	ch <- prometheus.MustNewConstMetric(c.location, prometheus.GaugeValue, 1, id, site)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
