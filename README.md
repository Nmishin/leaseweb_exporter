# Leaseweb Exporter

[![Docker](https://img.shields.io/badge/ghcr.io-nmishin%2Fleaseweb__exporter-blue?logo=docker)](https://github.com/Nmishin/leaseweb_exporter/pkgs/container/leaseweb_exporter)
[![License](https://img.shields.io/github/license/Nmishin/leaseweb_exporter)](LICENSE)

A [Prometheus](https://prometheus.io/) exporter for [Leaseweb](https://www.leaseweb.com/) dedicated servers. It exposes server metadata, location info, and hardware health status by querying the Leaseweb API.

## Features

- Exposes Prometheus metrics for Leaseweb dedicated servers
- Supports **multi-target scraping** — one exporter instance handles all your servers
- Built-in **HTTP Service Discovery** endpoint compatible with Prometheus `http_sd_configs`
- Discovery results are **cached for 1 hour** to minimize API calls
- Ships as a minimal, multi-arch Docker image (`linux/amd64`, `linux/arm64`)

## Metrics

The exporter provides detailed hardware telemetry and metadata for Leaseweb dedicated servers. It separates static metadata from dynamic health states to allow for efficient alerting and grouping.

### Server Information & Metadata
| Metric | Type | Labels | Description |
|---|---|---|---|
| `leaseweb_dedicated_server_info` | Gauge | `server_id`, `name`, `address` | Metadata about the server. Value is always `1`. |
| `leaseweb_dedicated_server_location` | Gauge | `server_id`, `site` | Physical location (Data Center) of the server. Value is always `1`. |

### Hardware Health Metrics
These metrics reflect the real-time status of the server's hardware components via IPMI.

| Metric | Type | Labels | Description |
|---|---|---|---|
| `leaseweb_dedicated_server_cooling_health` | Gauge | `server_id` | Health of the cooling system (fans, intake, exhaust). |
| `leaseweb_dedicated_server_drive_health` | Gauge | `server_id` | Health of the storage drives and backplane. |
| `leaseweb_dedicated_server_power_status` | Gauge | `server_id` | Current chassis power state. |

### Floating IP Metrics
Exposed on a separate endpoint, `/metrics/floating-ips`. Floating IPs belong to the account, not to one server, so they are scraped as a single target.

| Metric | Type | Labels | Description |
|---|---|---|---|
| `leaseweb_floating_ip_up` | Gauge | — | `1` if all Floating IPs API queries succeeded, `0` otherwise. |
| `leaseweb_floating_ip_info` | Gauge | `range_id`, `floating_ip`, `location`, `type`, `anchor_ip`, `passive_anchor_ip`, `status` | Floating IP definition and its current anchors. Value is always `1`. |
| `leaseweb_floating_ip_active` | Gauge | `range_id`, `floating_ip` | `1` if status is `ACTIVE`. |
| `leaseweb_floating_ip_updated_timestamp_seconds` | Gauge | `range_id`, `floating_ip` | Last change of the definition (e.g. anchor toggled). |
| `leaseweb_floating_ip_range_anchor_ips` | Gauge | `range_id` | Distinct anchor IPs currently in use in the range. |
| `leaseweb_floating_ip_range_expected_anchor_ips` | Gauge | `range_id` | `min(floating IPs, distinct anchor + passive IPs)` in the range. |
| `leaseweb_floating_ip_range_anchors_spread` | Gauge | `range_id` | `1` if floating IPs are spread across all anchor servers, `0` if they collapsed onto fewer servers (failover happened). |

---

## Health Status Reference

To simplify monitoring, the exporter uses a consistent **"1 = OK"** logic for all hardware sensors.

| Metric | Value | Meaning |
|---|---|---|
| **Cooling / Drive Health** | `1` | **Healthy** (Normal operation) |
| | `0` (or other) | **Fault** (Hardware error detected) |
| **Power Status** | `1` | **ON** (Server is running) |
| | `0` | **OFF** (Server is powered down) |

---

## HTTP Endpoints

| Endpoint | Description |
|---|---|
| `/metrics?target=<server_id>` | Prometheus metrics for a specific server |
| `/metrics/floating-ips` | Prometheus metrics for all Floating IP ranges of the account |
| `/targets` | HTTP Service Discovery — returns all servers as a target group |
| `/health` | Health check — returns `200 OK` |

## Configuration

The exporter is configured via environment variables:

| Variable | Required | Default | Description |
|---|---|---|---|
| `LW_EXPORTER_API_KEY` | **Yes** | — | Leaseweb API key |
| `LW_EXPORTER_ADDRESS` | No | `0.0.0.0` | Address to listen on |
| `LW_EXPORTER_PORT` | No | `9112` | Port to listen on |

You can generate a Leaseweb API key in the [Leaseweb Customer Portal](https://secure.leaseweb.com/api-client-management/).

## Getting Started

### Docker

```sh
docker run -d \
  -e LW_EXPORTER_API_KEY=your_api_key_here \
  -p 9112:9112 \
  ghcr.io/nmishin/leaseweb_exporter:latest
```

### Docker Compose

```yaml
services:
  leaseweb-exporter:
    image: ghcr.io/nmishin/leaseweb_exporter:latest
    restart: unless-stopped
    ports:
      - "9112:9112"
    environment:
      LW_EXPORTER_API_KEY: "your_api_key_here"
```

### Build from Source

Requires Go 1.21+.

```sh
git clone https://github.com/Nmishin/leaseweb_exporter.git
cd leaseweb_exporter
go build -o leaseweb-exporter ./cmd/leaseweb_exporter

LW_EXPORTER_API_KEY=your_api_key_here ./leaseweb-exporter
```

## Prometheus Configuration

This exporter follows the [multi-target exporter pattern](https://prometheus.io/docs/guides/multi-target-exporter/). Each server is scraped individually by passing its ID as the `target` query parameter.

### With HTTP Service Discovery (recommended)

The `/targets` endpoint returns all your servers in the [HTTP SD format](https://prometheus.io/docs/prometheus/latest/http_sd/), so Prometheus can discover them automatically.

```yaml
scrape_configs:
  - job_name: leaseweb
    http_sd_configs:
      - url: http://leaseweb-exporter:9112/targets
        refresh_interval: 1h
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: leaseweb-exporter:9112
```

### With a Static Target List

If you prefer to manage the server list manually:

```yaml
scrape_configs:
  - job_name: leaseweb
    static_configs:
      - targets:
          - "12345678"   # your Leaseweb server ID
          - "87654321"
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: leaseweb-exporter:9112
```

### Floating IPs

```yaml
scrape_configs:
  - job_name: leaseweb_floating_ips
    metrics_path: /metrics/floating-ips
    static_configs:
      - targets: ["leaseweb-exporter:9112"]
```

Example alerts:

```yaml
groups:
  - name: leaseweb-floating-ips
    rules:
      # Several floating IPs now point to the same server: the other one failed.
      - alert: LeasewebFloatingIPAnchorsCollapsed
        expr: leaseweb_floating_ip_range_anchors_spread == 0
        for: 5m
      # Anchor IP of a floating IP changed within the last 15 minutes.
      - alert: LeasewebFloatingIPAnchorChanged
        expr: count by (floating_ip) (max_over_time(leaseweb_floating_ip_info[15m])) > 1
      - alert: LeasewebFloatingIPExporterDown
        expr: leaseweb_floating_ip_up == 0
        for: 10m
```

### Verify a Single Scrape

```sh
curl "http://localhost:9112/metrics?target=12345678"
```

## License

[Apache 2.0 license](LICENSE)
