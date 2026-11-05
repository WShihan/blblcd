BINARY_NAME = blblcd
BUILD_DIR = bin
VERSION ?= $(shell git tag --sort=-creatordate | head -n 1 | sed 's/^[vV]//')
OPTIONS = CGO_ENABLED=0
COMMIT = $(shell git rev-parse HEAD)
BUILD_TIME = $(shell date +%Y-%m-%dT%H:%M:%S)
ENV = -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.Version=$(VERSION)
DARWIN_AMD = $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64
DARWIN_ARM = $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64
LINUX_AMD = $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64
LINUX_ARM = $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64
WIN_AMD = $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_windows_amd64
WIN_ARM = $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_windows_arm64

default: build deb

build:
	# 编译为 macOS 平台 amd64
	GOOS=darwin GOARCH=amd64 $(OPTIONS) go build -trimpath -ldflags="-s -w $(ENV)" -o $(DARWIN_AMD) main.go

	# 编译为 macOS 平台 arm64
	GOOS=darwin GOARCH=arm64 $(OPTIONS) go build -trimpath -ldflags="-s -w -w $(ENV)" -o $(DARWIN_ARM) main.go

	# 编译为 linux 平台 amd64
	GOOS=linux GOARCH=amd64 $(OPTIONS) go build -trimpath -ldflags="-s -w $(ENV)" -o $(LINUX_AMD) main.go

	# 编译为 linux 平台 arm64
	GOOS=linux GOARCH=arm64 $(OPTIONS)  go build -trimpath -ldflags="-s -w $(ENV)" -o $(LINUX_ARM)  main.go

	# 编译为 Windows 平台 amd64
	GOOS=windows GOARCH=amd64 $(OPTIONS) go build -trimpath -ldflags="-s -w $(ENV)" -o $(WIN_AMD).exe main.go

	# 编译为 Windows 平台 arm64
	GOOS=windows GOARCH=arm64 $(OPTIONS) go build -trimpath -ldflags="-s -w $(ENV)" -o $(WIN_ARM).exe main.go

debpkg_amd.yml: 
	@echo "name: $(BINARY_NAME)" > debpkg.yml
	@echo "version: $(VERSION)" >> debpkg.yml
	@echo "architecture: amd64" >> debpkg.yml
	@echo "maintainer: Shihan Wang" >> debpkg.yml
	@echo "maintainer_email: wangshihan751@gmail.com" >> debpkg.yml
	@echo "homepage: https://github.com/WShihan/blblcd" >> debpkg.yml
	@echo "description:" >> debpkg.yml
	@echo "  short: A minimalist blog system written in Go." >> debpkg.yml
	@echo "  long: |" >> debpkg.yml
	@echo "    A minimalist blog system written in Go." >> debpkg.yml
	@echo "files:" >> debpkg.yml
	@echo "  - file: $(LINUX_AMD)" >> debpkg.yml
	@echo "    dest: /usr/bin/$(BINARY_NAME)" >> debpkg.yml

debpkg_arm.yml: 
	@echo "name: $(BINARY_NAME)" > debpkg.yml
	@echo "version: $(VERSION)" >> debpkg.yml
	@echo "architecture: arm64" >> debpkg.yml
	@echo "maintainer: Shihan Wang" >> debpkg.yml
	@echo "maintainer_email: wangshihan751@gmail.com" >> debpkg.yml
	@echo "homepage: https://github.com/WShihan/blblcd" >> debpkg.yml
	@echo "description:" >> debpkg.yml
	@echo "  short: A minimalist blog system written in Go." >> debpkg.yml
	@echo "  long: |" >> debpkg.yml
	@echo "    A minimalist blog system written in Go." >> debpkg.yml
	@echo "files:" >> debpkg.yml
	@echo "  - file: $(LINUX_ARM)" >> debpkg.yml
	@echo "    dest: /usr/bin/$(BINARY_NAME)" >> debpkg.yml

deb_amd: debpkg_amd.yml
	debpkg -c debpkg.yml -v $(VERSION) -o $(LINUX_AMD).deb
	rm -rf debpkg.yml

deb_arm: debpkg_arm.yml
	debpkg -c debpkg.yml -v $(VERSION) -o $(LINUX_ARM).deb
	rm -rf debpkg.yml

deb: deb_amd deb_arm

clean:
	rm -rf $(BUILD_DIR)/$(BINARY_NAME)_*

.PHONY: build clean
