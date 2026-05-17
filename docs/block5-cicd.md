## Block 5: CI/CD and local development environment

### Scope

Block 5 is implemented around:

- `GitHub Actions self-hosted runner`
- `Kaniko` for container image builds
- local registry `spw-registry:5000`
- `Helm` chart for application services and gateway
- `Argo CD` refresh after chart update

### Why this path

The repository did not contain ready-made CI definitions or Helm charts. The least disruptive implementation path for the local lab is:

1. Keep the existing local `k3d + k3s + Argo CD` cluster.
2. Build images inside CI with `Kaniko`.
3. Push them into the registry that already belongs to the local `k3d` setup.
4. Update the Helm chart image tag in Git.
5. Ask `Argo CD` to refresh the application.

This closes the assignment flow without adding an external container registry or a second deployment mechanism.

### Implemented pieces

- Workflow: [block5-local-cicd.yml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/.github/workflows/block5-local-cicd.yml)
- Helm chart: [helm/smart-parking](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/helm/smart-parking)
- Kaniko build script: [build-push-kaniko.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/ci/build-push-kaniko.ps1)
- Helm image-tag updater: [update-helm-image-tag.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/ci/update-helm-image-tag.ps1)
- Argo refresh helper: [refresh-argocd-app.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/ci/refresh-argocd-app.ps1)

### Runner requirements

The self-hosted runner host must have:

- `Docker Desktop`
- `kubectl`
- `helm`
- access to the current local kubeconfig
- network access to the Docker network `k3d-spw-local`

The workflow expects the runner to execute on the same machine where the local cluster is running.

### Registry detail

The Helm chart uses `spw-registry:5000` as the in-cluster image registry reference.

This is intentional. Inside `k3d` nodes, the generated `registries.yaml` mirrors `spw-registry:5000` and `spw-registry:5001` to the registry container. For CI builds inside a Docker container, `spw-registry:5000` is the stable address to push to.

### Helm chart scope

The Helm chart currently packages:

- `smart-parking-app-config`
- `smart-parking-app-secret`
- `gateway-nginx-config`
- `user-service`
- `parking-service`
- `subscription-service`
- `reservation-service`
- `history-service`
- `notification-service`
- `gateway`

This satisfies the assignment requirement for Helm charts covering `3+` microservices and their infrastructure connection settings.

### Argo CD integration

`smart-parking-services` is switched from raw service manifests to the Helm chart path.

The CI pipeline applies the Argo application definition and then requests a hard refresh, so a new image tag committed into the chart is picked up by GitOps.

### What still depends on the local machine

- The runner itself still has to be installed and registered with GitHub manually.
- The workflow assumes `argocd`, `k3d`, and the local registry are already alive.
- The workflow updates the current branch, so for a production-like setup you would usually separate application source and deployment manifests into different repositories or branches.
