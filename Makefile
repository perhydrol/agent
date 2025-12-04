.PHONY: all build run test clean lint docker help swag

# 项目变量
APP_NAME=insurai
MAIN_FILE=cmd/main.go
# 你的 Docker 镜像名 (如果有的话)
DOCKER_IMAGE=$(APP_NAME):latest

# 默认目标
all: lint test build

# ------------------------------------
#  开发常用指令
# ------------------------------------

# 启动热重载开发 (依赖 Air)
run:
	@echo "Starting development server with Air..."
	air

# 编译二进制文件
build:
	@echo "Building binary..."
	go build -o tmp/$(APP_NAME) $(MAIN_FILE)

# 运行测试
test:
	@echo "Running unit tests..."
	go test -v ./...

# 生成 Swagger 文档
swag:
	@echo "Generating Swagger docs..."
	swag init -g $(MAIN_FILE) --output docs

# 代码格式化 (使用 gofmt)
fmt:
	@echo "Formatting code..."
	go fmt ./...

# ------------------------------------
#  代码质量检查
# ------------------------------------

# 运行 Linter (依赖 golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# ------------------------------------
#  辅助工具
# ------------------------------------

# 清理临时文件
clean:
	@echo "Cleaning up..."
	rm -rf tmp
	rm -rf docs

# 帮助信息
help:
	@echo "Makefile commands:"
	@echo "  make run    - Start dev server with hot reload (Air)"
	@echo "  make build  - Build binary to ./tmp"
	@echo "  make lint   - Run static analysis"
	@echo "  make swag   - Generate Swagger docs"
	@echo "  make test   - Run tests"
	@echo "  make fmt    - Format code"
	@echo "  make clean  - Remove tmp files"
