# Block 1.2: Local Node Autoscaling Decision

## Decision

`Karpenter` and `Cluster Autoscaler` are not being installed in the current local `k3d/k3s` lab.

For the local Block 1 scope, the repository uses:

- `k3d + k3s + Cilium` for the cluster itself
- controlled manual worker scaling through local scripts
- a documented rationale for why a cloud-style node autoscaler is deferred

## Why Karpenter Is Not Used Here

`Karpenter` is a node provisioning system that expects a supported cloud provider and infrastructure API behind the Kubernetes cluster.

For our local cluster:

- nodes are Docker containers managed by `k3d`
- there is no cloud provider API for provisioning machines
- there is no supported target environment similar to `EKS`

That makes `Karpenter` a poor fit for this lab stage.

## Why Cluster Autoscaler Is Not Used Here

`Cluster Autoscaler` scales node groups through provider-specific integration with the infrastructure layer.

For our local cluster:

- worker nodes are not managed by a cloud node group
- `k3d` does not expose an autoscaler-native machine group API
- introducing `Cluster API` just to make local node autoscaling work would add more platform complexity than value at this stage

## What We Use Instead

For local validation we need deterministic control over worker capacity, not cloud-grade elasticity.

This repository uses direct `k3d` node management commands for controlled local scaling.

PowerShell example:

```powershell
k3d node create spw-local-dynagent-2 --cluster spw-local --role agent --wait
```

Bash example:

```bash
k3d node create spw-local-dynagent-2 --cluster spw-local --role agent --wait
```

Remove an extra local worker:

```powershell
k3d node delete k3d-spw-local-dynagent-2-0
```

## Interpretation For Block 1

Block 1 is considered closed for the local lab if all of the following are true:

1. the cluster is reproducibly created with `k3d + k3s + Cilium`
2. Cilium network policy is validated with the smoke test
3. the operator has a clear runbook for switching from `docker compose` mode to `k3d/k8s` mode
4. worker-node capacity can be increased and decreased in a controlled way in the local environment
5. the autoscaling limitation is explicitly documented instead of hidden

## When To Revisit This

We should revisit real node autoscaling when the platform moves to one of these targets:

- a managed cloud Kubernetes cluster
- `Cluster API`-managed machines
- another environment with a supported node provisioning backend
