# Block 1: Local Platform Runbook

This runbook defines the working mode for Block 1 of the project.

## Short Answer

- Start `Docker Desktop`.
- Do not run the application with `docker compose` when working with `k3d/k3s`.
- Use `k3d + k3s + Cilium` as the only active runtime for Kubernetes tasks.
- Do not use `minikube` for this repository.
- If Docker Desktop Kubernetes is enabled, keep it disabled for this project to avoid confusion.

## Project Modes

There are two different local modes in this repository:

1. `docker compose` mode
2. `k3d/k3s` mode

Only one of them should be active at a time.

## docker compose Mode

Use this mode only when you want the old local stack from `docker-compose.yml`.

From the project root:

```powershell
docker compose up --build
```

Stop it from the same project root:

```powershell
docker compose down
```

Why from the root:

- `docker-compose.yml` is stored in the project root.
- Compose resolves the project and service names from that directory.

## k3d/k3s Mode

Use this mode for Block 1 and all further Kubernetes-related tasks.

Prepare the environment:

```powershell
.\scripts\k8s\prepare-local-k8s-mode.ps1
```

Create the cluster:

```powershell
.\scripts\k8s\create-local-cilium-cluster.ps1
```

Validate Cilium and network policies:

```powershell
.\scripts\k8s\validate-local-cilium-cluster.ps1
```

Delete the cluster when it is no longer needed:

```powershell
.\scripts\k8s\delete-local-cilium-cluster.ps1
```

## What Must Stay Running

- `Docker Desktop` must be running because `k3d` creates Kubernetes nodes as Docker containers.

## What Must Not Run At The Same Time

- `docker compose` application stack
- `minikube`
- Docker Desktop built-in Kubernetes for this repository

## Recommended Operator Flow

1. Start `Docker Desktop`.
2. Run `.\scripts\k8s\prepare-local-k8s-mode.ps1`.
3. Run `.\scripts\k8s\create-local-cilium-cluster.ps1`.
4. Run `.\scripts\k8s\validate-local-cilium-cluster.ps1`.
5. Continue with Kubernetes manifests and later blocks.

## Block 1 Status

### Task 1.1

Implemented in the repository:

- local `k3d` cluster config
- `k3s` runtime
- `Cilium` as CNI
- smoke validation for Cilium network policy

### Task 1.2

Closed for the local lab with a documented exception and local worker scaling workflow.

Use:

```powershell
k3d node create spw-local-dynagent-2 --cluster spw-local --role agent --wait
k3d node delete k3d-spw-local-dynagent-2-0
```

See the decision record:

- [block1-node-autoscaling-decision.md](/C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/docs/block1-node-autoscaling-decision.md)
