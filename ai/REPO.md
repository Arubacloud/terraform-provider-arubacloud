# REPO.md — Repository Organization

## Entry Point

`main.go` → `internal/provider/provider.go`

All provider logic lives in `internal/provider/`.

## Key Shared Files

- `internal/provider/resource_wait.go` — polling utility
- `internal/provider/error_helper.go` — API error formatting utility

## Resource Coverage

24 resources across domains:

| Domain | Resources |
|--------|-----------|
| Compute | CloudServer, Keypair, ElasticIP, Project |
| Storage | BlockStorage, Snapshot, Backup, Restore |
| Networking | VPC, Subnet, SecurityGroup, SecurityRule, VPCPeering, VPCPeeringRoute, VPNTunnel, VPNRoute |
| Container | ContainerRegistry, KaaS |
| Database | DBaaS, Database, DatabaseGrant, DatabaseBackup, DBaaSUser |
| Security | KMS |

Two resources (Key, KMIP) are **disabled** due to SDK limitations.
