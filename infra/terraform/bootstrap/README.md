# Terraform Bootstrap

This module bootstraps the base Kubernetes resources required by Block 2:

- namespaces
- base service accounts
- base application secret

## Resources

- `smart-parking` namespace
- `argocd` namespace
- `kafka` namespace
- `observability` namespace
- service accounts for runtime, migrations, and Argo CD bootstrap
- `smart-parking-app-secret`

## Prerequisites

- Terraform `>= 1.6`
- access to the target Kubernetes cluster through a kubeconfig file

## Usage

```bash
cd infra/terraform/bootstrap
terraform init
cp terraform.tfvars.example terraform.tfvars
terraform apply
```

## Important

The Kubernetes provider stores secret values in Terraform state. Treat the state file as sensitive.
