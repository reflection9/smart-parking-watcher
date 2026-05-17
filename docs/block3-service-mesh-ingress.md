# Block 3: Service Mesh, Ingress and Rate Limiting

## Selected stack

- `Istio` in sidecar mode for east-west traffic between `gateway` and backend services.
- `ingress-nginx` as the external Ingress Controller for local `k3d`.
- `Nginx gateway` remains the authoritative API Gateway and rate limiting layer.

## Why this split

`Istio` closes task `3.1` better than `Linkerd` for this project because it gives first-class `DestinationRule` and `VirtualService` primitives for:

- circuit breaking
- outlier detection
- retries and timeouts
- mTLS inside the namespace

For task `3.2`, `ingress-nginx` is simpler for local `k3d` than a separate `HAProxy + Keepalived` pair. The controller is exposed directly through `hostPort` on port `80/443`, while `k3d` already maps external traffic to the cluster on `localhost:8080` and `localhost:8443`.

## Rate limiting decision

Task `3.3` is implemented at the `gateway` level, not in `ingress-nginx`.

Reason:

- the project already has path-aware and method-aware limits in `gateway-nginx-config`
- some endpoints are write-heavy (`/users/register`, `/reservations`, `/history/archive`)
- coarse ingress-level retries or limits are less precise for these routes and can accidentally change behavior

Current enforced limits in the gateway:

- `POST /users/register` -> `5 req/min`
- `POST /users/login` -> `10 req/min`
- `POST /subscriptions` -> `20 req/min`
- `POST /reservations` and reservation state transitions -> `15 req/min`
- `POST /history/archive` -> `2 req/min`

## What Block 3 adds

- automatic sidecar injection for `smart-parking`
- strict namespace mTLS with `PeerAuthentication`
- `DestinationRule` policies for `gateway` and backend services
- safe read-only retries for selected HTTP `GET` traffic
- external `Ingress` that forwards all paths to `gateway`

## Validation targets

- `kubectl -n istio-system get pods`
- `kubectl -n ingress-nginx get pods`
- `kubectl -n argocd get applications.argoproj.io`
- `kubectl -n smart-parking get ingress`
