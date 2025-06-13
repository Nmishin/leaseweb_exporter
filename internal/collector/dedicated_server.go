package collector

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
        "github.com/Nmishin/leaseweb_exporter/internal/client"
)

type DedicatedServerCollector struct {
	servers  *prometheus.Desc
	location *prometheus.Desc
}

func NewDedicatedServerCollector() *DedicatedServerCollector {
	return &DedicatedServerCollector{
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

func (c *DedicatedServerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.servers
	ch <- c.location
}

func (c *DedicatedServerCollector) Collect(ch chan<- prometheus.Metric) {
	resp, _, err := client.LeasewebClient.DedicatedserverAPI.
		GetServerList(context.Background()).
		Execute()
	if err != nil {
		return // You may want to log this
	}

	servers := resp.GetServers()
	for _, server := range servers {
                id := deref(server.Id)
	        model := deref(server.Specs.Chassis)
	        site := deref(server.Location.Site)

		ch <- prometheus.MustNewConstMetric(
			c.servers,
			prometheus.GaugeValue,
			1,
			id,
			model,
		)

		ch <- prometheus.MustNewConstMetric(
			c.location,
			prometheus.GaugeValue,
			1,
			id,
			site,
		)
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
