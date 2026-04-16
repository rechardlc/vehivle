# vehivle 后端：常用命令入口（在 server/ 目录下执行：make <target>）
# 仓库根目录：make -C server <target>
# Windows：cmd 下可用；Git Bash / MSYS2 同样适用

# 默认 dev；与 configs.Load() 中 .env.{env} 一致
VEHIVLE_APP_ENV ?= dev

# 构建产物路径（Windows 下 go build -o 无扩展名时会生成 .exe）
ifeq ($(OS),Windows_NT)
  BIN := tmp/main.exe
else
  BIN := tmp/main
endif

.DEFAULT_GOAL := help

.PHONY: help deps tidy build run dev test test-unit fmt vet check clean migrate-up migrate-down migrate-version

## help: 显示本说明
help:
	@echo "Targets (run inside server/):"
	@echo "  make deps        - go mod download"
	@echo "  make tidy        - go mod tidy"
	@echo "  make build       - compile to $(BIN)"
	@echo "  make run         - VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV) go run ./cmd/api"
	@echo "  make dev         - air hot reload (requires air)"
	@echo "  make test        - go test ./..."
	@echo "  make test-unit   - go test ./tests/unit/..."
	@echo "  make fmt         - go fmt ./..."
	@echo "  make vet         - go vet ./..."
	@echo "  make check       - fmt + vet + test"
	@echo "  make clean       - delete build directory tmp/"
	@echo "  make migrate-up  - execute database migration up（requires VEHIVLE_DATABASE_DSN，cwd=server/）"
	@echo "  make migrate-down - rollback 1 step migration（-steps 1）"
	@echo "  make migrate-version - view current migration version"
	@echo "Env: VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV) (override: make run VEHIVLE_APP_ENV=prod)"

## deps: 拉取模块依赖
deps:
	go mod download

## tidy: 整理 go.mod / go.sum
tidy:
	go mod tidy

## build: 编译 API 二进制
ifeq ($(OS),Windows_NT)
build:
	@if not exist tmp mkdir tmp
	go build -o $(BIN) ./cmd/api
else
build:
	@mkdir -p tmp
	go build -o $(BIN) ./cmd/api
endif

## run: 直接运行（工作目录需为 server/；dev 时会加载 .env 与 .env.dev）
ifeq ($(OS),Windows_NT)
run:
	set VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV)&& go run ./cmd/api
else
run:
	VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV) go run ./cmd/api
endif

## dev: air 监听变更重启（见 .air.toml）
dev:
	air

## test: 全量测试
test:
	go test ./... -count=1

## test-unit: 单元测试目录
test-unit:
	go test ./tests/unit/... -count=1

## fmt: 格式化
fmt:
	go fmt ./...

## vet: 静态检查
vet:
	go vet ./...

## check: 提交前检查
check: fmt vet test

## clean: 清理本地构建产物
ifeq ($(OS),Windows_NT)
clean:
	-if exist tmp rmdir /s /q tmp
else
clean:
	rm -rf tmp
endif

## migrate-up / migrate-down / migrate-version：数据库迁移（工作目录需为 server/，与 make run 一致）
ifeq ($(OS),Windows_NT)
migrate-up:
	set VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV)&& go run ./cmd/migrate -op up
migrate-down:
	set VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV)&& go run ./cmd/migrate -op down -steps 1
migrate-version:
	set VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV)&& go run ./cmd/migrate -op version
else
migrate-up:
	VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV) go run ./cmd/migrate -op up
migrate-down:
	VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV) go run ./cmd/migrate -op down -steps 1
migrate-version:
	VEHIVLE_APP_ENV=$(VEHIVLE_APP_ENV) go run ./cmd/migrate -op version
endif
