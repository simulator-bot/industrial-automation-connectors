# Industrial Automation Connectors

**English** | [简体中文](README.zh-CN.md)

Connectors integrate PLCs, semiconductor equipment, robots, instruments, factory cameras, and other industrial systems. They expose common equipment, points, visual signals, events, and operations. Raw video travels over a dedicated media path. External vision services run visual algorithms.

## Design documents

- The [technical design](docs/architecture/technical-design.md) covers the Gateway, Process, and Frontend, the equipment capability model, protocols, subscriptions and commands, bounded rules, Linux and Kubernetes deployment, and rollout.
- Subsequent module designs belong under `docs/modules/`, including Gateway, protocol adapters, equipment models, rules runtime, and frontend.

The design-docs project owns system architecture and cross-project boundaries. This project owns connector technology, module interfaces, and implementation design.

## Starter scaffold

The repository now contains Go entrypoints for `cmd/gateway`, `cmd/process`,
and an isolated `cmd/rule-worker`, plus a Vue 3 + TypeScript + Vite operations
UI under `frontend/`. Shared equipment models, protocol adapter SPI, and v1
event/command contracts live under `internal/`; initial Kubernetes manifests
are under `deployments/kubernetes/base`.

Process currently uses in-memory equipment storage and exposes health checks and
an empty equipment list. Gateway has no concrete protocol adapters yet, and the
rule worker is idle pending the CEL runtime. No device collection or command
execution is claimed as implemented. Camera and vision integration are outside
this phase.

Run locally with `make run-process` and `make run-gateway` in separate
terminals. The frontend runs with `cd frontend && npm install && npm run dev`.
See `README.zh-CN.md` for the Chinese quick start and `docs/modules/README.md`
for module boundaries.
