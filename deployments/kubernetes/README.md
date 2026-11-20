# Kubernetes starter manifests

This is a deployment skeleton, not a production-ready release. Build separate
images with `docker build --build-arg SERVICE=process -t industrial-automation-connectors:process-dev .`,
`docker build --build-arg SERVICE=gateway -t industrial-automation-connectors:gateway-dev .`,
and `docker build -f frontend/Dockerfile -t industrial-automation-connectors:frontend-dev frontend/`
before applying the manifests.

The manifests intentionally use one replica. Gateway ownership leases, fencing,
protocol-specific network policy, credentials, persistence, and Gateway/Process
control transport must be implemented before enabling HA or production writes.
