# ==============================================================================
# Makefile for the wfpush project
# ==============================================================================

# 设置变量，方便以后修改
BINARY_NAME=wfpush

# --- 主命令 ---

# build: 编译 Go 程序
# "go build -o $(BINARY_NAME) ." 会编译当前目录下的代码，
# 并将生成的可执行文件命名为 wfpush。
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .
	@echo "$(BINARY_NAME) built successfully."

# run: 运行程序 (不带任何参数，启动服务模式)
run: build
	@echo "Running $(BINARY_NAME) service..."
	@./$(BINARY_NAME)

# clean: 清理所有生成的文件
# 这是您需要的核心功能。
clean:
	@echo "Cleaning up generated files..."
	# 使用 go clean 清理 Go 自己的构建产物
	@go clean
	# 使用 rm -f 强制删除我们程序运行时生成的文件
	# "-f" 参数意味着 "force"，即如果文件不存在也不会报错。
	@rm -f $(BINARY_NAME) config.yml data.json log.txt
	@echo "Cleanup complete."


# --- 辅助命令 ---

# .PHONY 告诉 make，这些目标（如 build, run, clean）是“伪目标”，
# 它们是命令的名称，而不是真实的文件名。
# 这是一个好习惯，可以避免与同名文件冲突。
.PHONY: build run clean