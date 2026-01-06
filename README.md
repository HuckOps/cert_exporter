# Certificate Exporter

A Prometheus Exporter written in Go for monitoring expiration time and validity of local and remote SSL/TLS certificates.

## Features

- 📡 **Remote Certificate Monitoring**: Check TLS certificates of remote servers via TCP connections
- 📁 **Local Certificate Monitoring**: Check certificate files on the local filesystem
- ⏰ **Asynchronous Scheduled Collection**: Background scheduled collection to avoid performance issues during Prometheus queries
- 📊 **Multi-dimensional Metrics**: Provide metrics for certificate expiry days, validity status, check latency, etc.
- 🛡️ **Thread-safe Cache**: Store collection results using a concurrent-safe caching mechanism

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/HuckOps/cert_exporter.git
cd cert_exporter

# Build the binary
go build -o cert_exporter

# Run
./cert_exporter --config config.yaml
```

### Run Directly

```bash
# Ensure Go 1.18+ is installed
go run . --config config.yaml
```

## Configuration

### Configuration File Example (config.yaml)

```yaml
# Log level: debug, info, warn, error
log_level: "info"

# Collection interval (seconds)
interval: 60

# Remote certificate monitoring targets
remote:
  - "example.com:443"
  - "github.com:443"
  - "prometheus.io:443"

# Local certificate monitoring targets
local:
  - public_key_path: "/path/to/local/cert.pem"
  - public_key_path: "/path/to/another/cert.pem"
```

### Command Line Arguments

| Argument | Description | Default Value |
|----------|-------------|---------------|
| `--config` | Path to configuration file | `config.yaml` |
| `--web.listen-address` | Address and port to listen on | `:9101` |
| `--web.telemetry-path` | Path under which to expose metrics | `/metrics` |

## Exposed Metrics

### `certificate_expiry_days`
Number of days until certificate expiry, Gauge type.

Labels:
- `domain`: Domain name corresponding to the certificate
- `sn`: Certificate serial number
- `source_type`: Certificate source (remote/local)
- `source`: Certificate source (configuration entry)

### `certificate_valid`
Whether the certificate is valid (1=valid, 0=invalid), Gauge type.

Labels:
- `domain`: Domain name corresponding to the certificate
- `sn`: Certificate serial number
- `source_type`: Certificate source (remote/local)
- `source`: Certificate source (configuration entry)

### `certificate_subject`
Certificate subject information, Gauge type.

Labels:
- `domain`: Domain name corresponding to the certificate
- `sn`: Certificate serial number
- `subject`: Certificate subject information
- `source_type`: Certificate source (remote/local)
- `source`: Certificate source (configuration entry)

### `certificate_check_status`
Certificate check status (1=success, 0=failure), Gauge type.

Labels:
- `domain`: Domain name corresponding to the certificate
- `source_type`: Certificate source (remote/local)
- `source`: Certificate source (configuration entry)

### `certificate_check_latency_milliseconds`
Time taken to check the certificate in milliseconds, Gauge type.

Labels:
- `domain`: Domain name corresponding to the certificate
- `source_type`: Certificate source (remote/local)
- `source`: Certificate source (configuration entry)

## Prometheus Configuration Example

Add the following content to your Prometheus configuration file:

```yaml
scrape_configs:
  - job_name: 'certificate_exporter'
    static_configs:
      - targets: ['localhost:9101']
    scrape_interval: 60s
```

## Technology Stack

- **Go**: 1.18+
- **Prometheus Client Library**: github.com/prometheus/client_golang
- **Zap**: go.uber.org/zap (logging framework)
- **YAML**: go.yaml.in/yaml/v2 (configuration parsing)

## License

[MIT License](LICENSE)
