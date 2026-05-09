# Observability Stack

## Chosen Stack

The project uses a composed observability stack:

- **OpenTelemetry instrumentation** inside the Go services
- **Jaeger** for distributed tracing
- **Prometheus** for metrics collection
- **Alertmanager** for alert routing
- **Grafana** for dashboards
- **Loki + Promtail** for centralized logs

## Why This Stack

This stack was chosen because it separates the three telemetry signals in a clear and explainable way:

- traces are handled by Jaeger
- metrics and alerts are handled by Prometheus + Alertmanager
- logs are collected by Loki + Promtail
- Grafana provides one UI layer over metrics, traces, and logs

For this project the split stack is easier to explain and verify than a heavier all-in-one platform. It also fits well with Docker-based local development and keeps every signal visible as an explicit architecture component.

## Comparison With Other Options

### SigNoz

SigNoz is a strong all-in-one observability platform built around OpenTelemetry with unified traces, metrics, logs, dashboards, and alerts. It is a very good product-level alternative, but for this project we preferred a more explicit stack where each component is visible separately in the architecture and easier to map to the report requirements.

### Jaeger

Jaeger is focused on distributed tracing. We use it specifically for that role instead of asking it to cover the whole observability surface.

### NetData

NetData is excellent for infrastructure-centric, real-time monitoring and quick anomaly detection. For our project it is less natural as the main observability center because we need clear application tracing and Prometheus-style alerting around our own microservices.

### Coroot

Coroot is a modern full-stack observability platform with strong eBPF-based capabilities and automatic coverage. It is a very interesting option, but for our project it would add more platform complexity than needed. We chose a simpler, more transparent stack that is easier to defend step by step.

## What Is Observable Now

- HTTP metrics for every service via `/metrics`
- distributed traces for incoming HTTP requests and inter-service HTTP calls
- centralized container logs through Loki/Promtail
- basic alert rules for:
  - service down
  - elevated 5xx rate
  - high p95 latency
