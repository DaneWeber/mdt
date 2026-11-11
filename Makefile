BINARY := mdt

.DEFAULT_GOAL := build

.PHONY: build install test clean help

build: ## Build the mdt binary in the repository root
	go build -o $(BINARY) .

install: ## Install mdt to $(go env GOPATH)/bin (or $GOBIN)
	go install .
	@case ":$$PATH:" in \
		*":$$(go env GOPATH)/bin:"*|*":$$(go env GOBIN):"*) ;; \
		*) echo "Note: add $$(go env GOPATH)/bin to your PATH to run $(BINARY) from anywhere." ;; \
	esac

test: ## Run the tests
	go test ./...

clean: ## Remove the built binary
	rm -f $(BINARY)

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  %-10s %s\n", $$1, $$2}'
