# pgoperator

A Toy Kubernetes operator for managing PostgreSQL database instances, built from scratch using raw `client-go` (no Kubebuilder/Operator SDK).
> ⚠️ this is strictly for learning purpose, don't take it seriously, WIP

## Overview

pgoperator watches for `PostgresDB` custom resources and automatically provisions and manages the full lifecycle of PostgreSQL instances in a Kubernetes cluster.

A single `PostgresDB` resource creates and manages:
- **StatefulSet** running PostgreSQL with persistent storage
- **Headless Service** for stable pod DNS (StatefulSet requirement)
- **ClusterIP Service** for client connections
- **Secret** with auto-generated credentials
- **PersistentVolumeClaims** per pod via StatefulSet volume claim templates

```yaml
apiVersion: bashketchum02.github.io/v1alpha1
kind: PostgresDB
metadata:
  name: my-postgres
spec:
  version: "16"
  replicas: 1
  storage: "1Gi"
  resources:
    requests:
      cpu: "250m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"
  backup:
    enabled: false
    schedule: "0 2 * * *"
    retentionDays: 7
```

## Architecture

```
kubectl apply PostgresDB
        │
        ▼
   API Server (etcd)
        │
   Watch (streaming)
        │
        ▼
   Informer ──► Event Handlers ──► Work Queue ──► Reconciler
                                                      │
                                          ┌───────────┼───────────┐
                                          ▼           ▼           ▼
                                       Secret     Services   StatefulSet
                                                                  │
                                                              PostgreSQL
                                                                Pods
```

The operator follows the standard Kubernetes controller pattern:
1. **Informer** watches the API server for `PostgresDB` resource changes
2. **Event handlers** extract the resource key and enqueue it
3. **Workers** pull keys from a rate-limited work queue
4. **Reconciler** compares desired state (spec) vs actual state (cluster) and creates/updates child resources

## Project Structure

```
pgoperator/
├── main.go                              # Entry point, K8s client setup
├── Makefile                             # Build, run, deploy targets
├── pkg/
│   ├── apis/postgresdb/v1alpha1/
│   │   ├── types.go                     # CRD Go types (PostgresDB, Spec, Status)
│   │   └── register.go                  # Scheme registration
│   └── controller/postgresdb/
│       ├── controller.go                # Informer, work queue, event handlers
│       ├── reconciler.go                # Ensure pattern for child resources
│       ├── statefulset.go               # StatefulSet construction
│       ├── service.go                   # Headless + client Service construction
│       └── secret.go                    # Credentials Secret + helpers
├── deploy/
│   ├── crds/postgresdb-crd.yaml         # CRD manifest
│   └── samples/sample-postgresdb.yaml   # Example PostgresDB resource
└── hack/
    └── kind-config.yaml                 # Local dev cluster config
```

## Prerequisites

- Go 1.21+
- Docker (or Colima)
- [kind](https://kind.sigs.k8s.io/)
- kubectl

## Quick Start

### 1. Create a local cluster

```bash
make cluster
```

### 2. Apply the CRD

```bash
make crd
```

### 3. Run the operator

```bash
make run
```

### 4. Create a PostgreSQL instance

```bash
make sample
```

### 5. Verify

```bash
kubectl get pgdb
kubectl get pods -l app.kubernetes.io/instance=my-postgres
kubectl get svc -l app.kubernetes.io/instance=my-postgres
kubectl get secret my-postgres-credentials
```

### 6. Connect to PostgreSQL

```bash
kubectl exec my-postgres-0 -- psql -U postgres -c "SELECT version();"
```

## CRD Spec Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `spec.version` | string | Yes | PostgreSQL version (`"14"`, `"15"`, `"16"`) |
| `spec.replicas` | int32 | Yes | Number of instances (1-10) |
| `spec.storage` | string | Yes | Storage per instance (e.g., `"10Gi"`) |
| `spec.resources` | object | No | CPU/memory requests and limits |
| `spec.backup.enabled` | bool | No | Enable scheduled backups |
| `spec.backup.schedule` | string | No | Cron expression for backup schedule |
| `spec.backup.retentionDays` | int32 | No | Days to retain backups (default: 7) |

## Key Dependencies

- `k8s.io/client-go` — Kubernetes Go client
- `k8s.io/apimachinery` — API types and machinery
- `k8s.io/api` — Built-in resource types
- `k8s.io/klog/v2` — Structured logging

## Status

This is a learning project built iteratively using raw `client-go` to understand Kubernetes API internals. Current status:

- [x] Phase 1: Project setup, local kind cluster
- [x] Phase 2: CRD types, scheme registration, YAML manifests
- [x] Phase 3: Controller with informers, work queue, reconciliation loop
- [x] Phase 4: Reconciler creating StatefulSet, Services, Secret
- [ ] Phase 5: Backup & Restore (CronJob for pg_dump, Job for pg_restore)
- [ ] Phase 6: Scaling, replication, credential rotation
- [ ] Phase 7: Monitoring sidecar, leader election, failover
