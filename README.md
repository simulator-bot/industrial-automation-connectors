# Industrial Automation Connectors

**English** | [简体中文](README.zh-CN.md)

Connectors integrate PLCs, semiconductor equipment, robots, instruments, factory cameras, and other industrial systems. They expose common equipment, points, visual signals, events, and operations. Raw video travels over a dedicated media path. External vision services run visual algorithms.

## Design documents

- The [technical design](docs/architecture/technical-design.md) covers the Gateway, Process, and Frontend, the equipment capability model, protocols, subscriptions and commands, bounded rules, Linux and Kubernetes deployment, and rollout.
- Subsequent module designs belong under `docs/modules/`, including Gateway, protocol adapters, equipment models, rules runtime, and frontend.

The design-docs project owns system architecture and cross-project boundaries. This project owns connector technology, module interfaces, and implementation design.
