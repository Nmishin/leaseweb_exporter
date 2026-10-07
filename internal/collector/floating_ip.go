package collector

import (
	"context"
	"log"
	"time"

	"github.com/Nmishin/leaseweb_exporter/internal/client"
	"github.com/prometheus/client_golang/prometheus"
)

type floatingIPLister interface {
	ListRanges(ctx context.Context) ([]client.FloatingIPRange, error)
	ListDefinitions(ctx context.Context, rangeID string) ([]client.FloatingIPDefinition, error)
}

type FloatingIPCollector struct {
	api floatingIPLister

	up                *prometheus.Desc
	info              *prometheus.Desc
	active            *prometheus.Desc
	updated           *prometheus.Desc
	anchorIPs         *prometheus.Desc
	expectedAnchorIPs *prometheus.Desc
	anchorsSpread     *prometheus.Desc
}

func NewFloatingIPCollector(api floatingIPLister) *FloatingIPCollector {
	return &FloatingIPCollector{
		api: api,

		up: prometheus.NewDesc(
			"leaseweb_floating_ip_up",
			"Whether the last Floating IPs API query succeeded (1=OK, 0=Fail)",
			nil, nil,
		),
		info: prometheus.NewDesc(
			"leaseweb_floating_ip_info",
			"Floating IP definition and its current anchor IPs",
			[]string{"range_id", "floating_ip", "location", "type", "anchor_ip", "passive_anchor_ip", "status"}, nil,
		),
		active: prometheus.NewDesc(
			"leaseweb_floating_ip_active",
			"Floating IP definition status (1=ACTIVE, 0=other)",
			[]string{"range_id", "floating_ip"}, nil,
		),
		updated: prometheus.NewDesc(
			"leaseweb_floating_ip_updated_timestamp_seconds",
			"Last time the floating IP definition was changed (e.g. anchor IP toggled)",
			[]string{"range_id", "floating_ip"}, nil,
		),
		anchorIPs: prometheus.NewDesc(
			"leaseweb_floating_ip_range_anchor_ips",
			"Number of distinct anchor IPs currently used by floating IPs in the range",
			[]string{"range_id"}, nil,
		),
		expectedAnchorIPs: prometheus.NewDesc(
			"leaseweb_floating_ip_range_expected_anchor_ips",
			"Number of distinct anchor IPs expected in the range: min(floating IPs, distinct anchor+passive IPs)",
			[]string{"range_id"}, nil,
		),
		anchorsSpread: prometheus.NewDesc(
			"leaseweb_floating_ip_range_anchors_spread",
			"Whether floating IPs in the range are spread across all anchor servers (1=OK, 0=collapsed onto fewer servers)",
			[]string{"range_id"}, nil,
		),
	}
}

func (c *FloatingIPCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.info
	ch <- c.active
	ch <- c.updated
	ch <- c.anchorIPs
	ch <- c.expectedAnchorIPs
	ch <- c.anchorsSpread
}

func (c *FloatingIPCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ranges, err := c.api.ListRanges(ctx)
	if err != nil {
		log.Printf("Could not list floating IP ranges: %v", err)
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		return
	}

	up := 1.0
	for _, r := range ranges {
		defs, err := c.api.ListDefinitions(ctx, r.ID)
		if err != nil {
			log.Printf("Could not list floating IP definitions for range %s: %v", r.ID, err)
			up = 0
			continue
		}
		c.collectRange(ch, r.ID, defs)
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up)
}

func (c *FloatingIPCollector) collectRange(ch chan<- prometheus.Metric, rangeID string, defs []client.FloatingIPDefinition) {
	anchors := map[string]struct{}{}
	pool := map[string]struct{}{}

	for _, d := range defs {
		ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1,
			rangeID, d.FloatingIP, d.Location, d.Type, d.AnchorIP, d.PassiveAnchorIP, d.Status)

		active := 0.0
		if d.Status == "ACTIVE" {
			active = 1
		}
		ch <- prometheus.MustNewConstMetric(c.active, prometheus.GaugeValue, active, rangeID, d.FloatingIP)

		if t, ok := parseLeasewebTime(d.UpdatedAt); ok {
			ch <- prometheus.MustNewConstMetric(c.updated, prometheus.GaugeValue, float64(t.Unix()), rangeID, d.FloatingIP)
		}

		if d.AnchorIP != "" {
			anchors[d.AnchorIP] = struct{}{}
			pool[d.AnchorIP] = struct{}{}
		}
		if d.PassiveAnchorIP != "" {
			pool[d.PassiveAnchorIP] = struct{}{}
		}
	}

	if len(defs) == 0 {
		return
	}

	// With N floating IPs over M servers, a healthy range uses min(N, M)
	// distinct anchors. Fewer means some server lost all its floating IPs,
	// i.e. a failover moved them onto another server.
	expected := min(len(defs), len(pool))
	spread := 0.0
	if len(anchors) >= expected {
		spread = 1
	}

	ch <- prometheus.MustNewConstMetric(c.anchorIPs, prometheus.GaugeValue, float64(len(anchors)), rangeID)
	ch <- prometheus.MustNewConstMetric(c.expectedAnchorIPs, prometheus.GaugeValue, float64(expected), rangeID)
	ch <- prometheus.MustNewConstMetric(c.anchorsSpread, prometheus.GaugeValue, spread, rangeID)
}

// The API returns both "+00:00" and "+0000" offsets.
func parseLeasewebTime(s string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05-0700"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
