package collector

import (
	"context"
	"strconv"
	"sync"

	"github.com/Nmishin/leaseweb_exporter/internal/client"
	"github.com/prometheus/client_golang/prometheus"
)

type DedicatedServerCollector struct {
	target    string
	collected *sync.Cond
	client    *client.Client

	servers  *prometheus.Desc
	location *prometheus.Desc
	health   *prometheus.Desc
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
		// 0 = OK, 1 = Warning, 2 = Critical, 3 = Unknown
		health: prometheus.NewDesc(
			"leaseweb_dedicated_server_health_status",
			"Hardware health status (0=OK, 1=Warning, 2=Critical, 3=Unknown)",
			[]string{"id"}, nil,
		),
	}
}

func (c *DedicatedServerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.servers
	ch <- c.location
	ch <- c.health
}

func (c *DedicatedServerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	resp, _, err := c.client.DedicatedserverAPI.
		GetServer(ctx, c.target).
		Execute()
	if err != nil {
		return
	}

	id := deref(resp.Id)
	model := deref(resp.Specs.Chassis)
	site := deref(resp.Location.Site)

	ch <- prometheus.MustNewConstMetric(c.servers, prometheus.GaugeValue, 1, id, model)
	ch <- prometheus.MustNewConstMetric(c.location, prometheus.GaugeValue, 1, id, site)

	healthResp, _, err := c.client.DedicatedserverAPI.
		GetHardwareMonitoring(ctx, id).
		Execute()

	if err == nil && healthResp != nil && len(healthResp.Metrics) > 0 {
		for _, m := range healthResp.Metrics {
			if m.Metric == "ipmi_current_state" {
				if val, parseErr := strconv.ParseFloat(m.Value, 64); parseErr == nil {
					ch <- prometheus.MustNewConstMetric(
						c.health,
						prometheus.GaugeValue,
						val,
						id,
					)
				}
				break
			}
		}
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
