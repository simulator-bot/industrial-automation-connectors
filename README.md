# Industrial Automation Connectors

**English** | [简体中文](README.zh-CN.md)

Connectors integrate PLCs, semiconductor equipment, robots, instruments, factory cameras, and other industrial systems. They expose common equipment, points, visual signals, events, and operations. Raw video travels over a dedicated media path. External vision services run visual algorithms.

## Design documents

- The [technical design](docs/architecture/technical-design.md) covers the Gateway, Process, and Frontend, the equipment capability model, protocols, subscriptions and commands, bounded rules, Linux and Kubernetes deployment, and rollout.
- Subsequent module designs belong under `docs/modules/`, including Gateway, protocol adapters, equipment models, rules runtime, and frontend.

The design-docs project owns system architecture and cross-project boundaries. This project owns connector technology, module interfaces, and implementation design.

## Current implementation

The repository now contains Go entrypoints for `cmd/gateway`, `cmd/process`,
and an isolated `cmd/rule-worker`, plus a Vue 3 + TypeScript + Vite operations
UI under `frontend/`. Shared equipment models, protocol adapter SPI, and v1
event/command contracts live under `internal/`; initial Kubernetes manifests
are under `deployments/kubernetes/base`.

Gateway can connect to a local OPC UA simulator, stream native monitored-item
notifications, and call configured methods. Process currently uses in-memory
equipment storage and exposes health checks and an empty equipment list. The
rule worker is idle pending the CEL runtime.

Run locally with `make run-process` and `make run-gateway` in separate
terminals. The frontend runs with `cd frontend && npm install && npm run dev`.
See `README.zh-CN.md` for the Chinese quick start and `docs/modules/README.md`
for module boundaries.

### Local OPC UA simulator

The [example configuration](deployments/examples/opcua-simulator.json) uses the
long-ingot inspection simulator on port 4840. Configure the endpoint, point
NodeIDs and method aliases for your simulator. The pilot uses one shared UA
session with anonymous authentication and SecurityPolicy None. Both the UA
endpoint and Gateway HTTP listener use loopback addresses. `localhost` in the
UA endpoint maps to IPv4. Use `[::1]` explicitly for an IPv6-only simulator.

```sh
GATEWAY_HTTP_ADDR=127.0.0.1:8081 \
GATEWAY_OPCUA_CONFIG=deployments/examples/opcua-simulator.json make run-gateway

# In another terminal, stream initial values and subsequent value/status changes.
curl -N http://127.0.0.1:8081/api/v1/gateway/opcua/events

# Write test text into the simulator's SetString method.
curl -fsS http://127.0.0.1:8081/api/v1/gateway/opcua/methods/set-string \
  -H 'Content-Type: application/json' \
  -d '{"arguments":[{"type":"UInt32","value":1},{"type":"UInt32","value":0},{"type":"String","value":"CONNECTORS-OPCUA-TEST"}]}'
```

Each SSE client opens an OPC UA data-change subscription for the configured
points and releases it on disconnect. Point events include the source timestamp
when provided, the ingestion timestamp, quality and raw OPC UA status code.
`intervalMillis` requests a publishing interval. Optional `deadband` sets an
absolute deadband on every configured point and requires numeric values supported
by the server's filter. After a terminal `subscription-error`, open a new stream.
If the UA session disconnects, restart Gateway before opening a new stream.
The pilot has no event replay.

Method arguments are positional, with explicit scalar UA types. Use
`{"arguments":[]}` for methods with no inputs. The response includes Call
status and typed output arguments. Interpret application-level results, such as
`Succeeded`, from the output arguments. The Gateway does not retry a Call.
Process orchestration, authorization and command history are future work.
