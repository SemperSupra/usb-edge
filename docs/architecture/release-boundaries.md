# Scope and release boundaries

- Read-only USB inventory and hardware observation
- Linux/ARM64 endpoint packaging, exporter lifecycle and driver restoration
- Fabric adapter consuming authoritative scoped grants

## Independent release qualification
Generic software release requires actual managed-exporter and restoration acceptance; Pi-branded hardware release separately requires real board qualification. Inventory/package PASS is not USB/IP or HIL PASS.

Support humans, automation and agents through stable CLI/API/schema, install/update/rollback instructions and typed errors. No new control plane or competing lease authority. Version contracts at the owning product.
