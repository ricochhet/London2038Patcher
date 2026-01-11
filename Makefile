BUILD_OUTPUT=build
ASSET_PATH=assets

CUSTOM=-X 'main.buildDate=$(shell date)' -X 'main.gitHash=$(shell git rev-parse --short HEAD)' -X 'main.buildOn=$(shell go version)'
LDFLAGS=$(CUSTOM) -w -s -extldflags=-static
GO_BUILD=go build -trimpath -ldflags "$(LDFLAGS)"

APP_NAMES=patcher fileserver

PATCHER_PATH=./cmd/london2038patcher
PATCHER_BIN_NAME=London2038Patcher

FILESERVER_PATH=./cmd/fileserver
FILESERVER_BIN_NAME=fileserver

define GO_BUILD_APP
	CGO_ENABLED=1 GOOS=$(1) GOARCH=$(2) $(GO_BUILD) -o $(BUILD_OUTPUT)/$(3) $(4)
endef

.PHONY: all
all: patcher fileserver

.PHONY: fmt
fmt:
	gofumpt -l -w -extra .

.PHONY: tidy
tidy:
	@echo "[main] tidy"
	go mod tidy

.PHONY: update
update:
	@echo "[main] update dependencies"
	go get -u ./...

.PHONY: lint
lint: fmt
	@echo "[main] golangci-lint"
	golangci-lint run ./... --fix

.PHONY: test
test:
	go test ./...

.PHONY: deadcode
deadcode:
	deadcode ./...

.PHONY: syso
syso:
	windres $(PATCHER_PATH)/app.rc -O coff -o $(PATCHER_PATH)/app.syso

.PHONY: png-to-icos
png-to-icos:
	magick $(ASSET_PATH)/win-icon.png -background none -define icon:auto-resize=256,128,64,48,32,16 $(ASSET_PATH)/win-icon.ico

# ----- Patcher -----
.PHONY: patcher
patcher: patcher-linux patcher-linux-arm64 patcher-darwin patcher-darwin-arm64 patcher-windows

.PHONY: patcher-linux
patcher-linux: fmt
	$(call GO_BUILD_APP,linux,amd64,$(PATCHER_BIN_NAME)-linux,$(PATCHER_PATH))

.PHONY: patcher-linux-arm64
patcher-linux-arm64: fmt
	$(call GO_BUILD_APP,linux,arm64,$(PATCHER_BIN_NAME)-linux-arm64,$(PATCHER_PATH))

.PHONY: patcher-darwin
patcher-darwin: fmt
	$(call GO_BUILD_APP,darwin,amd64,$(PATCHER_BIN_NAME)-darwin,$(PATCHER_PATH))

.PHONY: patcher-darwin-arm64
patcher-darwin-arm64: fmt
	$(call GO_BUILD_APP,darwin,arm64,$(PATCHER_BIN_NAME)-darwin-arm64,$(PATCHER_PATH))

.PHONY: patcher-windows
patcher-windows: fmt
	$(call GO_BUILD_APP,windows,amd64,$(PATCHER_BIN_NAME).exe,$(PATCHER_PATH))

# ----- FileServer -----
.PHONY: fileserver
fileserver: fileserver-linux fileserver-linux-arm64 fileserver-darwin fileserver-darwin-arm64 fileserver-windows

.PHONY: fileserver-linux
fileserver-linux: fmt
	$(call GO_BUILD_APP,linux,amd64,$(FILESERVER_BIN_NAME)-linux,$(FILESERVER_PATH))

.PHONY: fileserver-linux-arm64
fileserver-linux-arm64: fmt
	$(call GO_BUILD_APP,linux,arm64,$(FILESERVER_BIN_NAME)-linux-arm64,$(FILESERVER_PATH))

.PHONY: fileserver-darwin
fileserver-darwin: fmt
	$(call GO_BUILD_APP,darwin,amd64,$(FILESERVER_BIN_NAME)-darwin,$(FILESERVER_PATH))

.PHONY: fileserver-darwin-arm64
fileserver-darwin-arm64: fmt
	$(call GO_BUILD_APP,darwin,arm64,$(FILESERVER_BIN_NAME)-darwin-arm64,$(FILESERVER_PATH))

.PHONY: fileserver-windows
fileserver-windows: fmt
	$(call GO_BUILD_APP,windows,amd64,$(FILESERVER_BIN_NAME).exe,$(FILESERVER_PATH))
