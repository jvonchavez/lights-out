GOROOT_DIR := $(shell go env GOROOT)
WASM_OUT   := web/public/sim.wasm

# Colima exposes its Docker socket outside the default location, and its
# port forwarding does not reach testcontainers' Ryuk reaper. Tests clean up
# their own containers via t.Cleanup, so the reaper is redundant here.
# Harmless under Docker Desktop, where DOCKER_HOST is simply ignored.
COLIMA_SOCK := $(HOME)/.colima/default/docker.sock
DOCKER_ENV  := $(if $(wildcard $(COLIMA_SOCK)),DOCKER_HOST=unix://$(COLIMA_SOCK) TESTCONTAINERS_RYUK_DISABLED=true,)

.PHONY: help test test-integration test-all parity wasm balance simulate vet clean db db-down

help:
	@grep -hE '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  %-16s %s\n", $$1, $$2}'

test: ## Fast tests only: unit, property, golden (no Docker)
	go test -short ./... -count=1

test-integration: ## Integration tests against real Postgres in a container
	$(DOCKER_ENV) go test ./internal/store/... ./internal/api/... -count=1

test-all: parity ## Everything: unit, integration and native/WASM parity
	$(DOCKER_ENV) go test ./... -count=1

db: ## Start Postgres for local development
	$(DOCKER_ENV) docker compose up -d db

db-down: ## Stop and remove the development database
	$(DOCKER_ENV) docker compose down -v

vet: ## go vet
	go vet ./...

wasm: ## Build the js/wasm target and report its gzipped size
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(WASM_OUT) ./cmd/wasm
	cp "$(GOROOT_DIR)/lib/wasm/wasm_exec.js" web/public/
	@echo "sim.wasm: $$(wc -c < $(WASM_OUT)) bytes raw, $$(gzip -c $(WASM_OUT) | wc -c) bytes gzipped"

parity: wasm ## Assert native and js/wasm produce identical results
	go test -tags=parity ./internal/sim/ -run TestNativeWASMParity -v -count=1

balance: ## Run the 100k-season balance sweep
	go run ./cmd/balance -n 100000 -out balance.csv

simulate: ## Play one season from the CLI (SEED=1 STRATEGY=even)
	go run ./cmd/simulate -seed $${SEED:-1} -strategy $${STRATEGY:-aerofirst}

golden: ## Regenerate the golden fixtures (only after an intended rules change)
	go test ./internal/sim/ -run TestGolden -update -count=1

clean:
	rm -f $(WASM_OUT) web/public/wasm_exec.js balance*.csv
