GO ?= go
BIN_DIR ?= bin
SERVICES := process gateway rule-worker

.PHONY: fmt build run-process run-gateway run-rule-worker clean

fmt:
	$(GO) fmt ./...

build:
	@mkdir -p $(BIN_DIR)
	@for service in $(SERVICES); do \
		$(GO) build -o $(BIN_DIR)/$$service ./cmd/$$service || exit 1; \
	done

run-process:
	PROCESS_HTTP_ADDR="$${PROCESS_HTTP_ADDR:-127.0.0.1:8080}" \
	PROCESS_CATALOG_PATH="$${PROCESS_CATALOG_PATH:-var/equipment-catalog.json}" \
	PROCESS_GATEWAY_URL="$${PROCESS_GATEWAY_URL:-http://127.0.0.1:8081}" $(GO) run ./cmd/process

run-gateway:
	GATEWAY_HTTP_ADDR="$${GATEWAY_HTTP_ADDR:-127.0.0.1:8081}" $(GO) run ./cmd/gateway

run-rule-worker:
	$(GO) run ./cmd/rule-worker

clean:
	rm -rf $(BIN_DIR)
