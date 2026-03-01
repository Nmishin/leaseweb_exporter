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

| Metric | Type | Labels | Description |
|---|---|---|---|
| `leaseweb_dedicated_server_info` | Gauge | `server_id`, `name`, `address` | Metadata about a dedicated server. Always `1`. |
| `leaseweb_dedicated_server_location` | Gauge | `server_id`, `site` | Physical location of the server. Always `1`. |
| `leaseweb_dedicated_server_health_status` | Gauge | `server_id` | Hardware health status from IPMI monitoring. |

### Health Status Values

| Value | Meaning |
|---|---|
| `0` | OK |
| `1` | Warning |
| `2` | Critical |

## HTTP Endpoints

| Endpoint | Description |
|---|---|
| `/metrics?target=<server_id>` | Prometheus metrics for a specific server |
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

### Verify a Single Scrape

```sh
curl "http://localhost:9112/metrics?target=12345678"
```

## License

[Apache 2.0 license](LICENSE)
