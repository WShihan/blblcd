BINARY_NAME=blblcd
BUILD_DIR=build

default: build

build:
	# 编译为 Windows 平台
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe main.go

	# 编译为 macOS 平台
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 main.go

	# 编译为 Debian 平台
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 main.go

	# 编译为 BSD 平台
	GOOS=freebsd GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_freebsd_amd64 main.go

clean:
	rm -rf $(BUILD_DIR)

.PHONY: build clean