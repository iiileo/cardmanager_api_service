.PHONY: run-server dev test test-race generate-ent migrate-ent swagger fmt vet tidy

# Host may be Go 1.27+; pin toolchain to 1.25 for Ent generate compatibility.
export GOTOOLCHAIN ?= go1.25.0

run-server: ## Run API server
	go run ./cmd/server

dev: ## Hot reload with air
	air -c .air.toml

test: ## Run tests
	go test ./...

test-race: ## Run tests with race detector
	go test -race ./...

generate-ent: ## Generate Ent code from schema
	go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema

migrate-ent: ## Apply schema via server startup (ent.Schema.Create)
	@echo "Schema auto-migrates on server start (postgres.NewEntClient)"

swagger: ## Generate swagger docs
	swag init -g cmd/server/main.go -o docs/swagger

fmt: ## Format Go code
	gofmt -w .

vet: ## Vet Go code
	go vet ./...

tidy: ## Tidy modules and keep toolchain pin
	go mod tidy
	@grep -q '^toolchain ' go.mod || go mod edit -toolchain=go1.25.0
