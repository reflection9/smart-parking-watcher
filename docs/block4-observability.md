## Block 4: Observability

### Chosen stack

The Kubernetes observability stack for the local lab uses:

- `Prometheus` for metrics collection and alert evaluation
- `Alertmanager` for alert routing
- `Grafana` for dashboards and a single UI for metrics, logs, and traces
- `Loki + Promtail` for centralized container logs
- `Jaeger` for distributed tracing over OTLP

This keeps the Kubernetes deployment aligned with the existing Docker-based observability stack already present in the repository. It also stays lighter than a full Prometheus Operator based platform, which matters for `k3d + Docker Desktop`.

### Why this stack

Compared options from the assignment were narrowed down as follows:

- `Prometheus` was chosen over `VictoriaMetrics` and `InfluxDB` because the services already expose Prometheus-compatible `/metrics` endpoints and the alert rules were already written in PromQL.
- `Loki` was chosen over `VictoriaLogs` and `ELK` because the project already had Loki/Promtail configuration and Loki is materially lighter than Elasticsearch for a local lab.
- `Jaeger` was chosen over `Uptrace` and `Tempo` because the Go services are already wired for OTLP traces and the existing local stack already points them to Jaeger.
- `SigNoz` was not selected because it would replace the current split-stack approach with a heavier all-in-one platform, while the repo already has working per-signal building blocks.

### Deployed components

The GitOps layout for Block 4 is split in two parts:

- `infra/gitops/platform/04-jaeger.yaml`: Argo CD Application for the upstream Jaeger Helm chart
- `infra/gitops/platform/05-smart-parking-observability.yaml`: Argo CD Application for local manifests in `infra/k8s/platform/observability`

The local manifests deploy:

- Prometheus with Kubernetes pod discovery for annotated workloads
- Alertmanager with the existing default route
- Grafana with pre-provisioned datasources for Prometheus, Loki, and Jaeger
- Loki as a single-binary deployment with filesystem storage
- Promtail as a DaemonSet scraping Kubernetes pod logs
- Ingresses for Grafana and Prometheus

### Service instrumentation

Application services continue to expose metrics on `/metrics`. Distributed traces are sent to:

`jaeger.observability.svc.cluster.local:4318`

The value is injected through `smart-parking-app-config` and consumed by the shared Go observability package.

### Local access

With `ingress-nginx` running on host ports, the local UIs are available at:

- `http://grafana.localtest.me`
- `http://prometheus.localtest.me`
- `http://jaeger.localtest.me`

Default Grafana credentials:

- username: `admin`
- password: `admin`

### Operational notes

- This stack is intentionally ephemeral for the local lab. Prometheus, Loki, Alertmanager, and Grafana use `emptyDir` storage.
- `Promtail` is kept even though it is deprecated upstream because the repo already uses the Loki/Promtail model and the goal here is to close the assignment block with minimal architectural churn.
- Optional AI monitoring tooling is intentionally out of scope for this block.
