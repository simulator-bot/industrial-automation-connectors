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
	$(GO) run ./cmd/process

run-gateway:
	$(GO) run ./cmd/gateway

run-rule-worker:
	$(GO) run ./cmd/rule-worker

clean:
	rm -rf $(BIN_DIR)
