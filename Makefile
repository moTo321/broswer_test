# 自动化测试框架 Makefile

# 变量定义
BINARY_NAME=browse_test
MAIN_PACKAGE=main.go
BUILD_DIR=build
ASSETS_DIR=assets
GO_VERSION=1.21

# 默认目标
.PHONY: all
all: build

# 编译项目
.PHONY: build
build:
	@echo "🔨 编译项目..."
	@mkdir -p $(BUILD_DIR)
	@if go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE) 2>&1 | grep -q "leptonica\|tesseract"; then \
		echo "⚠️  检测到 Tesseract OCR 依赖问题，尝试使用无验证码模式编译..."; \
		go build -tags nocaptcha -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE); \
		echo "✅ 编译完成（无验证码功能）: $(BUILD_DIR)/$(BINARY_NAME)"; \
		echo "   提示: 如需验证码识别功能，请运行: make install-tesseract"; \
	else \
		echo "✅ 编译完成: $(BUILD_DIR)/$(BINARY_NAME)"; \
	fi

# 编译项目（无验证码功能）
.PHONY: build-nocaptcha
build-nocaptcha:
	@echo "🔨 编译项目（无验证码功能）..."
	@mkdir -p $(BUILD_DIR)
	@go build -tags nocaptcha -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✅ 编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 编译（Windows）
.PHONY: build-windows
build-windows:
	@echo "🔨 编译 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	@if GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PACKAGE) 2>&1 | grep -q "leptonica\|tesseract"; then \
		echo "⚠️  使用无验证码模式编译..."; \
		GOOS=windows GOARCH=amd64 go build -tags nocaptcha -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PACKAGE); \
		echo "✅ 编译完成（无验证码功能）: $(BUILD_DIR)/$(BINARY_NAME).exe"; \
	else \
		echo "✅ 编译完成: $(BUILD_DIR)/$(BINARY_NAME).exe"; \
	fi

# 编译（Linux）
.PHONY: build-linux
build-linux:
	@echo "🔨 编译 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	@if GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PACKAGE) 2>&1 | grep -q "leptonica\|tesseract"; then \
		echo "⚠️  使用无验证码模式编译..."; \
		GOOS=linux GOARCH=amd64 go build -tags nocaptcha -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PACKAGE); \
		echo "✅ 编译完成（无验证码功能）: $(BUILD_DIR)/$(BINARY_NAME)-linux"; \
	else \
		echo "✅ 编译完成: $(BUILD_DIR)/$(BINARY_NAME)-linux"; \
	fi

# 编译（macOS）
.PHONY: build-macos
build-macos:
	@echo "🔨 编译 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	@if GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-macos $(MAIN_PACKAGE) 2>&1 | grep -q "leptonica\|tesseract"; then \
		echo "⚠️  使用无验证码模式编译..."; \
		GOOS=darwin GOARCH=amd64 go build -tags nocaptcha -o $(BUILD_DIR)/$(BINARY_NAME)-macos $(MAIN_PACKAGE); \
		echo "✅ 编译完成（无验证码功能）: $(BUILD_DIR)/$(BINARY_NAME)-macos"; \
	else \
		echo "✅ 编译完成: $(BUILD_DIR)/$(BINARY_NAME)-macos"; \
	fi

# 编译所有平台
.PHONY: build-all
build-all: build-windows build-linux build-macos
	@echo "✅ 所有平台编译完成"

# 运行项目（使用 go run）
.PHONY: run
run:
	@echo "🚀 运行项目..."
	@go run $(MAIN_PACKAGE)

# 运行项目（使用编译后的二进制文件）
.PHONY: run-bin
run-bin: build
	@echo "🚀 运行编译后的程序..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# 运行测试用例（指定文件）
.PHONY: test
test:
	@echo "🧪 运行测试用例..."
	@go run $(MAIN_PACKAGE) -f $(TEST_FILE)

# 运行测试用例（使用编译后的二进制文件）
.PHONY: test-bin
test-bin: build
	@echo "🧪 运行测试用例..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -f $(TEST_FILE)

# 安装依赖
.PHONY: deps
deps:
	@echo "📦 安装依赖..."
	@if [ ! -f go.mod ]; then \
		echo "初始化 Go 模块..."; \
		go mod init autotest; \
	fi
	@go mod download
	@go mod tidy
	@echo "✅ 依赖安装完成"

# 安装 Playwright
.PHONY: install-playwright
install-playwright:
	@echo "📦 安装 Playwright..."
	@go install github.com/playwright-community/playwright-go/cmd/playwright@latest
	@playwright install chromium
	@echo "✅ Playwright 安装完成"

# 安装 Tesseract OCR（验证码识别需要）
.PHONY: install-tesseract
install-tesseract:
	@echo "📦 安装 Tesseract OCR..."
	@if command -v apt-get > /dev/null; then \
		echo "检测到 Debian/Ubuntu 系统，使用 apt-get 安装..."; \
		sudo apt-get update && sudo apt-get install -y tesseract-ocr libtesseract-dev; \
	elif command -v yum > /dev/null; then \
		echo "检测到 CentOS/RHEL 系统，使用 yum 安装..."; \
		sudo yum install -y tesseract tesseract-devel; \
	elif command -v brew > /dev/null; then \
		echo "检测到 macOS 系统，使用 brew 安装..."; \
		brew install tesseract; \
	else \
		echo "⚠️  无法自动检测系统类型，请手动安装 Tesseract OCR"; \
		echo "   Debian/Ubuntu: sudo apt-get install tesseract-ocr libtesseract-dev"; \
		echo "   CentOS/RHEL: sudo yum install tesseract tesseract-devel"; \
		echo "   macOS: brew install tesseract"; \
		exit 1; \
	fi
	@echo "✅ Tesseract OCR 安装完成"

# 格式化代码
.PHONY: fmt
fmt:
	@echo "📝 格式化代码..."
	@go fmt ./...
	@echo "✅ 代码格式化完成"

# 代码检查
.PHONY: vet
vet:
	@echo "🔍 代码检查..."
	@go vet ./...
	@echo "✅ 代码检查完成"

# 代码检查（使用 golangci-lint）
.PHONY: lint
lint:
	@echo "🔍 代码检查 (golangci-lint)..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过检查"; \
	fi

# 清理编译文件
.PHONY: clean
clean:
	@echo "🧹 清理编译文件..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✅ 清理完成"

# 清理所有生成的文件（包括 assets）
.PHONY: clean-all
clean-all: clean
	@echo "🧹 清理所有生成的文件..."
	@rm -rf $(ASSETS_DIR)/errors/*
	@rm -rf $(ASSETS_DIR)/captcha/*
	@rm -rf $(ASSETS_DIR)/videos/*
	@echo "✅ 清理完成"

# 创建必要的目录
.PHONY: init-dirs
init-dirs:
	@echo "📁 创建必要的目录..."
	@mkdir -p $(ASSETS_DIR)/errors
	@mkdir -p $(ASSETS_DIR)/captcha
	@mkdir -p $(ASSETS_DIR)/videos
	@mkdir -p $(BUILD_DIR)
	@echo "✅ 目录创建完成"

# 初始化项目（安装依赖、创建目录等）
.PHONY: init
init: init-dirs deps install-playwright
	@echo "✅ 项目初始化完成"
	@echo ""
	@echo "⚠️  注意: 如果需要使用验证码识别功能，请运行:"
	@echo "   make install-tesseract"

# 显示帮助信息
.PHONY: help
help:
	@echo "自动化测试框架 - Makefile 命令"
	@echo ""
	@echo "编译命令:"
	@echo "  make build          - 编译当前平台版本"
	@echo "  make build-windows  - 编译 Windows 版本"
	@echo "  make build-linux    - 编译 Linux 版本"
	@echo "  make build-macos    - 编译 macOS 版本"
	@echo "  make build-all      - 编译所有平台版本"
	@echo ""
	@echo "运行命令:"
	@echo "  make run            - 使用 go run 运行"
	@echo "  make run-bin        - 运行编译后的二进制文件"
	@echo "  make test TEST_FILE=testcase/login_example.json  - 运行指定测试用例"
	@echo "  make test-bin TEST_FILE=testcase/login_example.json  - 使用编译后的程序运行测试"
	@echo ""
	@echo "开发命令:"
	@echo "  make deps           - 安装 Go 依赖"
	@echo "  make install-playwright  - 安装 Playwright"
	@echo "  make fmt            - 格式化代码"
	@echo "  make vet            - 代码检查"
	@echo "  make lint           - 代码检查 (golangci-lint)"
	@echo ""
	@echo "清理命令:"
	@echo "  make clean          - 清理编译文件"
	@echo "  make clean-all      - 清理所有生成的文件"
	@echo ""
	@echo "初始化命令:"
	@echo "  make init           - 初始化项目（创建目录、安装依赖等）"
	@echo "  make init-dirs      - 创建必要的目录"
	@echo ""
	@echo "示例:"
	@echo "  make build && make run-bin"
	@echo "  make test TEST_FILE=testcase/login_example.json"
	@echo "  make build-all"

