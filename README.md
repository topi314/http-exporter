[![Go Report](https://goreportcard.com/badge/github.com/topi314/prometheus-exporters)](https://goreportcard.com/report/github.com/topi314/prometheus-exporters)
[![Go Version](https://img.shields.io/github/go-mod/go-version/topi314/prometheus-exporters)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/github/license/topi314/prometheus-exporters)](LICENSE)
[![Version](https://img.shields.io/github/v/tag/topi314/prometheus-exporters?label=release)](https://github.com/topi314/prometheus-exporters/releases/latest)
[![Docker](https://github.com/topi314/prometheus-exporters/actions/workflows/build.yml/badge.svg)](https://github.com/topi314/prometheus-exporters/actions/workflows/build.yml)
[![Discord](https://discordapp.com/api/guilds/608506410803658753/embed.png?style=shield)](https://discord.gg/sD3ABd5)

# prometheus-exporters

This repository contains a collection of random Prometheus exporters I use.

## Installation

You can either run the exporters directly or use the provided Docker image.

### Docker-Compose

```yaml
services:
  http-exporter:
    image: ghcr.io/topi314/http-exporter:master
    container_name: http-exporter
    restart: unless-stopped
    volumes:
      - ./config.toml:/var/lib/http-exporter/config.toml
    ports:
      - "2112:2112"
```

## Configuration

The exporters are configured via a TOML file. The default path is `/var/lib/http-exporter/config.toml` but you can change it with the `--config` flag.

```toml
[global]
scrape_interval = "1m"
scrape_timeout = "10s"

[log]
level = "info"
format = "text"
add_source = false

[server]
listen_addr = ":2112"
endpoint = "/metrics"

# Add your exporter configurations here
# [[configs]]
# name = "Bla"
# type = "http-temp"
# interval = "1m"
# timeout = "10s"
# [configs.options]
```

## Exporters

### HTTP Temperature Exporter

This exporter reads temperature data from a HTTP endpoint and exposes it as a Prometheus gauge metric.

#### Configuration

```toml
[[configs]]
name = "Bla"
type = "http-temp"
interval = "1m"
timeout = "10s"

[configs.options]
# The metric name, help text and labels
metric = { name = "bla_temp", help = "Temperature in celsius", labels = { name = "bla" } }

# The HTTP endpoint to fetch the data from
address = "hostname:port"
insecure = true
username = "user"
password = "password"
```

## License

Shelly Exporter is licensed under the [Apache License 2.0](LICENSE).

## Contributing

Contributions are always welcome! Just open a pull request or discussion and I will take a look at it.

## Contact

- [Discord](https://discord.gg/sD3ABd5)
- [Twitter](https://twitter.com/topi3141)
- [Email](mailto:hi@topi.wtf)
