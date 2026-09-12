.PHONY: run-server test test-race generate-ent migrate-ent generate-migration swagger fmt vet tidy

run-server: ## Run API server
	go run ./cmd/server

test: ## Run tests
	go test ./...

test-race: ## Run tests with race detector
	go test -race ./...

generate-ent: ## Generate Ent code from schema
	go generate ./ent

migrate-ent: ## Apply Ent schema to local DB (requires schemas)
	@echo "Add ent schemas first, then wire migrate-ent"

generate-migration: ## Export SQL migrations
	@echo "Add ent schemas first, then wire generate-migration"

swagger: ## Generate swagger docs
	swag init -g cmd/server/main.go -o docs/swagger

fmt: ## Format Go code
	gofmt -w .

vet: ## Vet Go code
	go vet ./...

tidy: ## Tidy modules
	go mod tidy
