.DEFAULT_GOAL := help

.PHONY: proto
proto: check-deps
	@echo "🚀 Launch script for Protocol Buffer code generation..."
	@chmod +x scripts/generate-proto.sh
	@./scripts/generate-proto.sh

.PHONY: proto-deps
proto-deps:
	@echo "🛠️ Installing protobuf dependencies..."
	@which protoc > /dev/null || (echo "⚠️  Note: protoc not found. Please install protobuf-compiler" && sleep 2)
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: check-deps
check-deps:
	@echo "🔍 Checking build dependencies..."
	@which protoc > /dev/null || (echo "❌ Error: protoc not installed.\n   On Ubuntu: sudo apt-get install protobuf-compiler\n   On macOS: brew install protobuf" && exit 1)
	@[ -f "$(shell go env GOPATH)/bin/protoc-gen-go" ] || (echo "❌ Error: protoc-gen-go not installed. Run 'make proto-deps'" && exit 1)
	@[ -f "$(shell go env GOPATH)/bin/protoc-gen-go-grpc" ] || (echo "❌ Error: protoc-gen-go-grpc not installed. Run 'make proto-deps'" && exit 1)
	@echo "✅ All dependencies are satisfied"

.PHONY: all
all: proto-deps proto
	@echo "✅ Build setup completed!"

.PHONY: clean-proto
clean-proto:
	@echo "🧹 Cleaning generated protobuf code..."
	@rm -rf pkg/proto/*

.PHONY: swag
swag: check-swag-deps
	@echo "🚀 Launch script for Swagger documentation generation..."
	@chmod +x scripts/generate-swag.sh
	@./scripts/generate-swag.sh

.PHONY: swag-deps
swag-deps:
	@echo "🛠️ Installing Swaggo dependencies..."
	@go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: check-swag-deps
check-swag-deps:
	@echo "🔍 Checking Swaggo dependencies..."
	@which swag > /dev/null || (echo "❌ Error: swag not installed. Run 'make swag-deps'" && exit 1)
	@echo "✅ Swaggo dependencies are satisfied"

.PHONY: test
test:
	@go test ./...

.PHONY: build
build: proto swag
	@go build ./...

.PHONY: docker-down-v
docker-down-v:
	@echo "🐳 Stopping and removing containers with volume cleanup..."
	@docker compose down -v

.PHONY: docker-up
docker-up:
	@echo "🐳 Starting services with Docker Compose..."
	@docker compose up -d

.PHONY: migrate
migrate:
	@echo "🗄️ Running database migrations..."
	@chmod +x scripts/migrate.sh
	@./scripts/migrate.sh

.PHONY: setup-db
setup-db: docker-down-v docker-up
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@./scripts/migrate.sh up

.PHONY: cleanup-db
cleanup-db:
	@echo "🗄️ Cleaning up database (rolling back migrations)..."
	@chmod +x scripts/migrate.sh
	@./scripts/migrate.sh down

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  proto         - Generate protobuf code"
	@echo "  proto-deps    - Install Go protobuf dependencies (requires protoc)"
	@echo "  swag          - Generate Swagger documentation"
	@echo "  swag-deps     - Install Swaggo dependencies"
	@echo "  all           - Install proto-deps and generate proto"
	@echo "  test          - Run tests"
	@echo "  clean-proto   - Remove generated protobuf code"
	@echo "  docker-down-v - Stop containers and remove volumes"
	@echo "  docker-up     - Start services with Docker Compose"
	@echo "  migrate       - Run database migrations"
	@echo "  cleanup-db    - Rollback all database migrations"
	@echo "  setup-db      - Clean start with Docker and run migrations"
	@echo ""
	@echo "⚠️ Note: protoc must be installed separately via system package manager"
