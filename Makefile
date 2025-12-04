CUSTOM=-X 'main.buildDate=$(shell date)' -X 'main.gitHash=$(shell git rev-parse --short HEAD)' -X 'main.buildOn=$(shell go version)'
LDFLAGS=$(CUSTOM) -w -s -extldflags=-static

GO_BUILD=go build -trimpath -ldflags "$(LDFLAGS)"
BUILD_OUTPUT=build
ASSET_PATH=assets

APP_PATH=./cmd/london2038patcher

.PHONY: all
all: london2038patcher-linux london2038patcher-linux-arm64 london2038patcher-darwin london2038patcher-darwin-arm64 london2038patcher-windows

.PHONY: all-windows
all-windows: london2038patcher-windows

.PHONY: fmt
fmt:
	gofumpt -l -w -extra .

.PHONY: tidy
tidy:
#	go get -u ./...
	@echo "[main] tidy"
	go mod tidy

.PHONY: update
update:
	@echo "[main] tidy"
	go get -u ./...

.PHONY: lint
lint: fmt
# golangci-lint cache clean
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
	windres $(APP_PATH)/app.rc -O coff -o $(APP_PATH)/app.syso

.PHONY: png-to-icos
png-to-icos:
	magick $(ASSET_PATH)/win-icon.png -background none -define icon:auto-resize=256,128,64,48,32,16 $(ASSET_PATH)/win-icon.ico

.PHONY: london2038patcher-linux
london2038patcher-linux: fmt
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/london2038patcher-linux $(APP_PATH)

.PHONY: london2038patcher-linux-arm64
london2038patcher-linux-arm64: fmt
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(BUILD_OUTPUT)/london2038patcher-linux-arm64 $(APP_PATH)

.PHONY: london2038patcher-darwin
london2038patcher-darwin: fmt
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/london2038patcher-darwin $(APP_PATH)

.PHONY: london2038patcher-darwin-arm64
london2038patcher-darwin-arm64: fmt
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(BUILD_OUTPUT)/london2038patcher-darwin-arm64 $(APP_PATH)

.PHONY: london2038patcher-windows
london2038patcher-windows: fmt
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/London2038Patcher.exe $(APP_PATH)