# Classius Development Makefile
# Complete workflow for building the classical education platform

.PHONY: help setup clean build test deploy docs

# Default target
help:
	@echo "Classius Development Commands:"
	@echo ""
	@echo "Setup & Environment:"
	@echo "  make setup          - Initial development environment setup"
	@echo "  make install-deps   - Install all dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  make dev            - Start full development environment"
	@echo "  make device-sim     - Launch device simulator"
	@echo "  make server-dev     - Start development server"
	@echo "  make ai-dev         - Start AI services in development mode"
	@echo ""
	@echo "Building:"
	@echo "  make build          - Build all components"
	@echo "  make build-device   - Build device firmware"
	@echo "  make build-server   - Build server components"
	@echo "  make build-release  - Build production release"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-device    - Run device tests"
	@echo "  make test-server    - Run server tests"
	@echo "  make test-ai        - Run AI service tests"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy-dev     - Deploy to development environment"
	@echo "  make deploy-prod    - Deploy to production"
	@echo "  make flash-device   - Flash device firmware"
	@echo ""
	@echo "Documentation:"
	@echo "  make docs           - Generate documentation"
	@echo "  make docs-serve     - Serve documentation locally"

# Variables
DEVICE_DIR = src/device
SERVER_DIR = src/server
AI_DIR = src/server/ai
SHARED_DIR = src/shared
BUILD_DIR = build
DOCS_DIR = docs

# Cross-compilation setup
ARM_TOOLCHAIN = arm-linux-gnueabihf-
export CROSS_COMPILE=$(ARM_TOOLCHAIN)

###########################################
# SETUP & ENVIRONMENT
###########################################

setup: install-deps init-git setup-device setup-server
	@echo "✅ Classius development environment ready!"

install-deps:
	@echo "📦 Installing dependencies..."
	@./scripts/install-dependencies.sh

init-git:
	@echo "🔧 Setting up git hooks..."
	@cp tools/hooks/* .git/hooks/
	@chmod +x .git/hooks/*

setup-device:
	@echo "🖥️  Setting up device development..."
	@cd $(DEVICE_DIR) && qmake6 -project
	@cd $(DEVICE_DIR) && qmake6

setup-server:
	@echo "🖧 Setting up server development..."
	@cd $(SERVER_DIR) && go mod init github.com/classius/server
	@cd $(SERVER_DIR) && go mod tidy

###########################################
# DEVELOPMENT
###########################################

dev: dev-services device-sim
	@echo "🚀 Development environment started!"

dev-services:
	@echo "🔧 Starting development services..."
	@docker-compose -f docker/docker-compose.dev.yml up -d
	@cd $(AI_DIR) && python -m uvicorn main:app --reload --port 8001 &
	@cd $(SERVER_DIR) && air & # Hot reload for Go

device-sim:
	@echo "📱 Starting device simulator..."
	@cd $(DEVICE_DIR) && ./classius-simulator

server-dev:
	@echo "🖧 Starting development server..."
	@cd $(SERVER_DIR) && go run main.go

ai-dev:
	@echo "🤖 Starting AI services..."
	@cd $(AI_DIR) && python -m uvicorn main:app --reload --host 0.0.0.0 --port 8001

###########################################
# BUILDING
###########################################

build: build-device build-server
	@echo "✅ Build complete!"

build-device:
	@echo "📱 Building device firmware..."
	@mkdir -p $(BUILD_DIR)/device
	@cd $(DEVICE_DIR) && qmake6 CONFIG+=release
	@cd $(DEVICE_DIR) && make
	@cp $(DEVICE_DIR)/classius $(BUILD_DIR)/device/

build-device-arm:
	@echo "📱 Building device firmware for ARM..."
	@mkdir -p $(BUILD_DIR)/device-arm
	@cd $(DEVICE_DIR) && qmake6 CONFIG+=arm-cross CONFIG+=release
	@cd $(DEVICE_DIR) && make
	@cp $(DEVICE_DIR)/classius $(BUILD_DIR)/device-arm/

build-server:
	@echo "🖧 Building server..."
	@mkdir -p $(BUILD_DIR)/server
	@cd $(SERVER_DIR) && go build -o ../../$(BUILD_DIR)/server/classius-server ./cmd/server
	@cd $(AI_DIR) && python -m pip install -r requirements.txt

build-release: clean build-device-arm build-server
	@echo "📦 Creating release package..."
	@mkdir -p $(BUILD_DIR)/release
	@cp -r $(BUILD_DIR)/device-arm/* $(BUILD_DIR)/release/
	@cp -r $(BUILD_DIR)/server/* $(BUILD_DIR)/release/
	@tar -czf $(BUILD_DIR)/classius-release-$(shell date +%Y%m%d).tar.gz -C $(BUILD_DIR)/release .

###########################################
# TESTING
###########################################

test: test-device test-server test-ai
	@echo "✅ All tests passed!"

test-device:
	@echo "🧪 Running device tests..."
	@cd $(DEVICE_DIR) && make test

test-server:
	@echo "🧪 Running server tests..."
	@cd $(SERVER_DIR) && go test ./...

test-ai:
	@echo "🧪 Running AI service tests..."
	@cd $(AI_DIR) && python -m pytest tests/

test-integration:
	@echo "🧪 Running integration tests..."
	@python -m pytest tests/integration/

###########################################
# DEPLOYMENT
###########################################

deploy-dev: build
	@echo "🚀 Deploying to development environment..."
	@./scripts/deploy-dev.sh

deploy-prod: build-release
	@echo "🚀 Deploying to production..."
	@./scripts/deploy-prod.sh

flash-device:
	@echo "⚡ Flashing device firmware..."
	@./scripts/flash-device.sh $(BUILD_DIR)/device-arm/classius

###########################################
# DOCUMENTATION
###########################################

docs:
	@echo "📚 Generating documentation..."
	@cd $(DEVICE_DIR) && doxygen Doxyfile
	@cd $(SERVER_DIR) && go doc ./...
	@cd $(AI_DIR) && python -m sphinx.cmd.build docs docs/_build

docs-serve:
	@echo "📖 Serving documentation at http://localhost:8080"
	@python -m http.server 8080 --directory $(DOCS_DIR)

###########################################
# UTILITIES
###########################################

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@cd $(DEVICE_DIR) && make clean || true
	@cd $(SERVER_DIR) && go clean

format:
	@echo "🎨 Formatting code..."
	@cd $(DEVICE_DIR) && clang-format -i **/*.cpp **/*.h
	@cd $(SERVER_DIR) && go fmt ./...
	@cd $(AI_DIR) && black . && isort .

lint:
	@echo "🔍 Linting code..."
	@cd $(SERVER_DIR) && golangci-lint run
	@cd $(AI_DIR) && flake8 . && mypy .

# Database operations
db-setup:
	@echo "🗄️  Setting up database..."
	@docker-compose -f docker/docker-compose.dev.yml exec postgres psql -U classius -c "CREATE DATABASE IF NOT EXISTS classius_dev;"

db-migrate:
	@echo "🗄️  Running database migrations..."
	@cd $(SERVER_DIR) && go run cmd/migrate/main.go

db-seed:
	@echo "🌱 Seeding database with test data..."
	@cd $(SERVER_DIR) && go run cmd/seed/main.go

# Container operations
docker-build:
	@echo "🐳 Building Docker images..."
	@docker-compose -f docker/docker-compose.yml build

docker-up:
	@echo "🐳 Starting Docker services..."
	@docker-compose -f docker/docker-compose.yml up -d

docker-down:
	@echo "🐳 Stopping Docker services..."
	@docker-compose -f docker/docker-compose.yml down

# Quick development shortcuts
quick-test: test-server test-ai
	@echo "⚡ Quick tests complete!"

rebuild: clean build
	@echo "🔄 Rebuild complete!"

restart-dev:
	@docker-compose -f docker/docker-compose.dev.yml restart
	@echo "🔄 Development services restarted!"

# Show current project status
status:
	@echo "📊 Classius Project Status:"
	@echo ""
	@echo "📁 Structure:"
	@find . -name "*.go" -o -name "*.cpp" -o -name "*.h" -o -name "*.py" -o -name "*.qml" | head -20
	@echo ""
	@echo "🗄️  Database:"
	@docker-compose -f docker/docker-compose.dev.yml exec postgres psql -U classius -c "SELECT 'Connected to PostgreSQL' as status;" 2>/dev/null || echo "Database not running"
	@echo ""
	@echo "🐳 Containers:"
	@docker-compose -f docker/docker-compose.dev.yml ps

.DEFAULT_GOAL := help