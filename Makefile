# 辅助工具安装（请手动执行一次）：
# go install github.com/cloudwego/hertz/cmd/hz@latest
# go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
# go install github.com/hertz-contrib/swagger-generate/thrift-gen-http-swagger@latest

.DEFAULT_GOAL := help

MODULE = github.com/FantasyRL/go-mcp-demo
REMOTE_REPOSITORY ?= fantasyrl/go-mcp-demo
DIR = $(abspath .)
CMD = $(DIR)/cmd
CONFIG_PATH = $(DIR)/config
IDL_PATH = $(DIR)/idl
GEN_CONFIG_PATH ?= $(DIR)/pkg/gorm-gen/generator/etc/config.yaml

DOCKER_NET := docker_mcp_net
IMAGE_PREFIX ?= hachimi
TAG          ?= $(shell git rev-parse --short HEAD 2>nul || echo dev)

SERVICES := host mcp_local mcp_remote
service = $(word 1, $@)

# --- 代码生成 ---
.PHONY: hertz-gen-api
hertz-gen-api:
	hz update -idl "$(IDL_PATH)/api.thrift"
	powershell -NoProfile -Command " \
		if (Test-Path '$(DIR)/swagger') { Remove-Item -Recurse -Force '$(DIR)/swagger' }; \
		if (Test-Path '$(DIR)/gen-go') { Remove-Item -Recurse -Force '$(DIR)/gen-go' } \
	"
	thriftgo -g go -p http-swagger "$(IDL_PATH)/api.thrift"

# --- 运行本地服务 ---
.PHONY: $(SERVICES)
$(SERVICES):
	go run "$(CMD)/$(service)" -cfg "$(CONFIG_PATH)/config.yaml"

# --- 数据库模型生成 ---
.PHONY: model
model:
	@echo Generating database models...
	go run "$(DIR)/pkg/gorm-gen/generator" -f "$(GEN_CONFIG_PATH)"

# --- Vendor 依赖 ---
.PHONY: vendor
vendor:
	@echo ">> go mod tidy && go mod vendor"
	go mod tidy
	go mod vendor

# --- 构建 Docker 镜像 ---
.PHONY: docker-build-%
docker-build-%: vendor
	@echo ">> Building image for service: $* (tag: $(TAG))"
	docker build ^
	  --build-arg SERVICE=$* ^
	  -f docker/Dockerfile ^
	  -t $(IMAGE_PREFIX)/$*:$(TAG) ^
	  .

# --- 拉取并运行容器（使用 PowerShell 脚本）---
.PHONY: pull-run-%
pull-run-%:
	@echo ">> Pulling and running docker (Windows): $*"
	docker pull $(REMOTE_REPOSITORY):$*
	powershell -NoProfile -ExecutionPolicy Bypass -File "$(DIR)\scripts\docker-run.ps1" -Service "$*" -Image "$(REMOTE_REPOSITORY):$*" -ConfigPath "$(CONFIG_PATH)\config.yaml"

# --- 帮助信息 ---
.PHONY: help
help:
	@echo Available targets:
	@echo   host                 - Run cmd/host with config.yaml
	@echo   mcp_local           - Run cmd/mcp_local with config.yaml
	@echo   mcp_remote          - Run cmd/mcp_remote with config.yaml
	@echo   vendor              - Run 'go mod tidy && go mod vendor'
	@echo   model               - Generate GORM models
	@echo   hertz-gen-api       - Regenerate Hertz API from IDL
	@echo   docker-build-^<svc^> - Build Docker image for service
	@echo   pull-run-^<svc^>    - Pull and run container via PowerShell script
	@echo   stdio               - Build mcp_local.exe and run host with stdio config
	@echo   env                 - Start Consul + dependencies via docker-compose
	@echo   push-^<svc^>        - Push image to remote repo (with confirmation)

# --- Stdio 模式（用于 MCP 测试）---
.PHONY: stdio
stdio:
	@echo ">> Building mcp_local.exe for stdio mode..."
	go build -o bin/mcp_local.exe ./cmd/mcp_local
	@echo ">> Running host with stdio config..."
	go run ./cmd/host -cfg "$(CONFIG_PATH)/config.stdio.yaml"

# --- 推送镜像（带确认）---
.PHONY: push-%
push-%:
	powershell -NoProfile -Command " \
		$$svc = '$*'; \
		$$validServices = @('host', 'mcp_local', 'mcp_remote'); \
		if ($$validServices -notcontains $$svc) { \
			Write-Host 'ERROR: Service $$svc is not valid. Available: [$$($$validServices -join ', ')]' -ForegroundColor Red; \
			exit 1 \
		}; \
		$$confirm = Read-Host 'Confirm service name to push (type ''$$svc'' to confirm)'; \
		if ($$confirm -ne $$svc) { \
			Write-Host 'Confirmation failed. Expected ''$$svc'', got ''$$confirm''' -ForegroundColor Red; \
			exit 1 \
		}; \
		$$arch = (Get-CimInstance Win32_Processor).Architecture; \
		Write-Host 'Building and pushing image for service: $$svc'; \
		if ($$arch -eq 9 -or $$arch -eq 6) { \
			docker build --build-arg SERVICE=$$svc -t $(REMOTE_REPOSITORY):$$svc -f docker/Dockerfile .; \
			docker push $(REMOTE_REPOSITORY):$$svc; \
		} else { \
			Write-Host 'Using buildx for cross-platform build...'; \
			docker buildx build --platform linux/amd64 --build-arg SERVICE=$$svc -t $(REMOTE_REPOSITORY):$$svc -f docker/Dockerfile --push .; \
		} \
	"

# --- 启动开发环境（Consul 等）---
.PHONY: env
env:
	powershell -NoProfile -Command " \
		$$consulPath = '$(DIR)/docker/data/consul'; \
		if (Test-Path $$consulPath) { \
			Write-Host 'Removing old Consul data...'; \
			Remove-Item -Recurse -Force $$consulPath; \
		}; \
		Set-Location '$(DIR)/docker'; \
		docker-compose up -d \
	"

# --- CI/CD 专用推送（无交互）---
.PHONY: push-cd-%
push-cd-%: vendor
	powershell -NoProfile -Command " \
		$$svc = '$*'; \
		$$validServices = @('host', 'mcp_local', 'mcp_remote'); \
		if ($$validServices -notcontains $$svc) { \
			Write-Host 'ERROR: Invalid service $$svc' -ForegroundColor Red; \
			exit 1 \
		}; \
		$$arch = (Get-CimInstance Win32_Processor).Architecture; \
		if ($$arch -eq 9 -or $$arch -eq 6) { \
			docker build --build-arg SERVICE=$$svc -t $(REMOTE_REPOSITORY):$$svc -f docker/Dockerfile .; \
			docker push $(REMOTE_REPOSITORY):$$svc; \
		} else { \
			docker buildx build --platform linux/amd64 --build-arg SERVICE=$$svc -t $(REMOTE_REPOSITORY):$$svc -f docker/Dockerfile --push .; \
		} \
	"