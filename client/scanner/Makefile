# Makefile for Blockchain Block Scanner

# 变量定义
BINARY_NAME=block-scanner
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# 默认目标
.PHONY: all
all: clean build

# 构建
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/main.go
	@echo "Build complete: ${BUILD_DIR}/${BINARY_NAME}"

# 运行
.PHONY: run
run:
	@echo "Running ${BINARY_NAME}..."
	go run ./cmd/main.go

# 开发模式运行
.PHONY: dev
dev:
	@echo "Running ${BINARY_NAME} in development mode..."
	go run ./cmd/main.go --config config/config.yaml

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# 测试覆盖率
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "Checking code..."
	go vet ./...

# 代码检查（包含更多工具）
.PHONY: lint
lint: fmt vet
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# 清理
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf ${BUILD_DIR}
	@rm -f coverage.out coverage.html
	@go clean

# 安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 更新依赖
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# 生成文档
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060; \
	else \
		echo "godoc not found, skipping..."; \
	fi

# 创建发布版本
.PHONY: release
release: clean build
	@echo "Creating release..."
	@mkdir -p release
	@cp ${BUILD_DIR}/${BINARY_NAME} release/${BINARY_NAME}-${VERSION}
	@cp config/config.yaml release/config.yaml.example
	@cp README.md release/
	@echo "Release created: release/"

# 交叉编译
.PHONY: cross-build
cross-build:
	@echo "Cross compiling..."
	@mkdir -p ${BUILD_DIR}
	
	# Linux
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-arm64 ./cmd/main.go
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 ./cmd/main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 ./cmd/main.go
	
	# Windows
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe ./cmd/main.go
	
	@echo "Cross compilation complete"

# 安装到系统
.PHONY: install
install: build
	@echo "Installing ${BINARY_NAME}..."
	@sudo cp ${BUILD_DIR}/${BINARY_NAME} /usr/local/bin/
	@echo "Installation complete"

# 卸载
.PHONY: uninstall
uninstall:
	@echo "Uninstalling ${BINARY_NAME}..."
	@sudo rm -f /usr/local/bin/${BINARY_NAME}
	@echo "Uninstallation complete"

# 扫描示例
.PHONY: scan-btc
scan-btc:
	@echo "Scanning Bitcoin blocks continuously..."
	./${BUILD_DIR}/${BINARY_NAME} --chain btc --start 800000

.PHONY: scan-eth
scan-eth:
	@echo "Scanning Ethereum blocks continuously..."
	./${BUILD_DIR}/${BINARY_NAME} --chain eth --start 18000000

.PHONY: scan-all
scan-all:
	@echo "Scanning all chains continuously..."
	./${BUILD_DIR}/${BINARY_NAME} --interval 60s

# 帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  run           - Run the application"
	@echo "  dev           - Run in development mode"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  fmt           - Format code"
	@echo "  vet           - Check code"
	@echo "  lint          - Lint code"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  docs          - Generate documentation"
	@echo "  release       - Create release package"
	@echo "  cross-build   - Cross compile for multiple platforms"
	@echo "  install       - Install to system"
	@echo "  uninstall     - Uninstall from system"
	@echo "  scan-btc      - Scan Bitcoin blocks"
	@echo "  scan-eth      - Scan Ethereum blocks"
	@echo "  scan-all      - Scan all chains"
	@echo "  help          - Show this help"
