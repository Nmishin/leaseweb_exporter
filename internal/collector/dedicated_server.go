package collector

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

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

	discoveryCache []string
	lastDiscovery  time.Time
	cacheMutex     sync.RWMutex
}

func NewDedicatedServerCollector(target string) *DedicatedServerCollector {
	return &DedicatedServerCollector{
		target:    target,
		client:    &client.LeasewebClient,
		collected: sync.NewCond(&sync.Mutex{}),

		servers: prometheus.NewDesc(
			"leaseweb_dedicated_server_info",
			"Metadata about a Leaseweb dedicated server",
			[]string{"server_id", "name", "address"}, nil,
		),
		location: prometheus.NewDesc(
			"leaseweb_dedicated_server_location",
			"Server location info",
			[]string{"server_id", "site"}, nil,
		),
		health: prometheus.NewDesc(
			"leaseweb_dedicated_server_health_status",
			"Hardware health status (0=OK, 1=Warning, 2=Critical, 3=Unknown)",
			[]string{"server_id"}, nil,
		),
	}
}

func (c *DedicatedServerCollector) GetAllServerIDs(ctx context.Context) ([]string, error) {
	c.cacheMutex.RLock()
	if c.discoveryCache != nil && time.Since(c.lastDiscovery) < 5*time.Minute {
		log.Printf("Returning cached discovery results (age: %v, count: %d)",
			time.Since(c.lastDiscovery).Round(time.Second),
			len(c.discoveryCache))
		c.cacheMutex.RUnlock()
		return c.discoveryCache, nil
	}
	c.cacheMutex.RUnlock()

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	if time.Since(c.lastDiscovery) < 5*time.Minute && c.discoveryCache != nil {
		return c.discoveryCache, nil
	}

	log.Printf("Starting discovery for all dedicated servers via API...")

	var allIDs []string
	var currentOffset int32 = 0
	const apiMaxLimit int32 = 50

	for {
		req := c.client.DedicatedserverAPI.GetServerList(ctx).
			Limit(apiMaxLimit).
			Offset(currentOffset)

		serverResponse, _, err := req.Execute()
		if err != nil {
			return nil, fmt.Errorf("retrieving the list of servers: %w", err)
		}

		if serverResponse.Servers != nil {
			for _, s := range serverResponse.Servers {
				allIDs = append(allIDs, deref(s.Id))
			}
		}

		if len(serverResponse.Servers) < int(apiMaxLimit) {
			break
		}

		currentOffset += apiMaxLimit
	}

	c.discoveryCache = allIDs
	c.lastDiscovery = time.Now()

	log.Printf("Discovery finished. Found %d servers.", len(allIDs))
	return allIDs, nil
}

func (c *DedicatedServerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.servers
	ch <- c.location
	ch <- c.health
}

func (c *DedicatedServerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	resp, _, err := c.client.DedicatedserverAPI.GetServer(ctx, c.target).Execute()
	if err != nil {
		return
	}

	id := deref(resp.Id)
	site := deref(resp.Location.Site)
	name := deref(resp.Contract.Reference.Get())

	address := "unknown"
	ipResp, _, err := c.client.DedicatedserverAPI.GetIpList(ctx, id).NetworkType("PUBLIC").Execute()
	if err == nil && ipResp.Ips != nil {
		for _, ipObj := range ipResp.Ips {
			if ipObj.MainIp != nil && *ipObj.MainIp {
				rawIP := deref(ipObj.Ip)
				address = strings.Split(rawIP, "/")[0]
				break
			}
		}
		if address == "unknown" && len(ipResp.Ips) > 0 {
			rawIP := deref(ipResp.Ips[0].Ip)
			address = strings.Split(rawIP, "/")[0]
		}
	}

	ch <- prometheus.MustNewConstMetric(c.servers, prometheus.GaugeValue, 1, id, name, address)
	ch <- prometheus.MustNewConstMetric(c.location, prometheus.GaugeValue, 1, id, site)

	healthResp, _, err := c.client.DedicatedserverAPI.GetHardwareMonitoring(ctx, id).Execute()
	if err == nil && healthResp != nil && len(healthResp.Metrics) > 0 {
		for _, m := range healthResp.Metrics {
			if m.Metric == "ipmi_current_state" {
				if val, parseErr := strconv.ParseFloat(m.Value, 64); parseErr == nil {
					ch <- prometheus.MustNewConstMetric(c.health, prometheus.GaugeValue, val, id)
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
