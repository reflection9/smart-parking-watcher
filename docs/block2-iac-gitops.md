# Block 2: IaC and GitOps

This block adds three capabilities to the project:

1. Terraform bootstrap for base cluster resources
2. Argo CD installation and App of Apps bootstrap
3. Ansible role for Strimzi-backed Kafka deployment

## Terraform Bootstrap

Path:

```text
infra/terraform/bootstrap
```

Creates:

- namespaces: `smart-parking`, `argocd`, `kafka`, `observability`
- service accounts for runtime, migrations, and Argo CD bootstrap
- `smart-parking-app-secret`

Run:

```bash
cd infra/terraform/bootstrap
terraform init
cp terraform.tfvars.example terraform.tfvars
terraform apply
```

## Argo CD

GitOps paths:

```text
infra/gitops/apps
infra/k8s/app/infra
infra/k8s/app/migrations
infra/k8s/app/services
```

Bootstrap scripts:

```text
scripts/k8s/bootstrap-argocd.ps1
scripts/k8s/bootstrap-argocd.sh
```

PowerShell:

```powershell
.\scripts\k8s\bootstrap-argocd.ps1
```

The bootstrap script:

- creates the `argocd` namespace
- installs Argo CD from the official stable manifest
- renders the root Application with the current Git remote URL and current branch
- applies the root Application

## App of Apps Layout

Root Application:

- `smart-parking-root`

Child Applications:

- `smart-parking-infra`
- `smart-parking-migrations`
- `smart-parking-services`

All three use automated sync with `prune` and `selfHeal`.

At the moment, the child Applications track the working Git branch:

```text
codex/block2-iac-gitops
```

After merging this work, switch `targetRevision` to the branch or tag used by the target environment.

## Ansible Role For Kafka

Paths:

```text
ansible/playbooks/deploy-strimzi-kafka.yml
ansible/roles/strimzi_kafka
```

Role behavior:

- ensures the `kafka` namespace exists
- installs the Strimzi Cluster Operator from the official install URL
- waits for the operator to become available
- applies a Kafka custom resource
- waits until the Kafka resource reports `Ready=True`

## Notes

- Terraform and Ansible binaries must be installed locally to execute those parts.
- Terraform state will contain raw secret values.
- The current Argo CD child app for migrations applies one-shot Kubernetes Jobs from the repository. For repeated migration runs, delete the old migration Jobs before re-syncing or convert them to a dedicated hook strategy later.
