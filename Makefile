GOROOT_DIR := $(shell go env GOROOT)
WASM_OUT   := web/public/sim.wasm

.PHONY: help test test-all parity wasm balance simulate vet clean

help:
	@grep -hE '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  %-16s %s\n", $$1, $$2}'

test: ## Unit, property and golden tests (no Docker required)
	go test ./... -count=1

test-all: ## Everything, including Docker-backed integration tests
	go test ./... -count=1
	go test -tags=parity ./internal/sim/ -count=1

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
