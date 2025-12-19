# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOLINT=bin/golangci-lint run
BINARY_NAME=krb5tray
MODULE_NAME=krb5tray
BUILD_DIR=bin

ifeq ($(OS),Windows_NT)
  RMDIR = if exist "$(1)\*" rmdir /S /Q "$(1)"
else
  RMDIR = rm -rf -- "$(1)"
endif

DATE=$(shell date +%Y%m%d_%H%M%S)
COMMIT=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags="-X main.commit=$(COMMIT) -X main.buildDate=$(DATE)"

-include Makefile.local

# Compilation targets
all: darwin-arm64 windows-amd64
#all: darwin-arm64 linux-amd64

darwin-arm64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o=$(BUILD_DIR)/$(BINARY_NAME).darwin-arm64.bin 
	codesign -s "SorinS_Signing" bin/krb5tray.darwin-arm64.bin

darwin-amd64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).darwin-amd64.bin

linux-amd64: 
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).linux-amd64.bin

linux-arm64:
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).linux-arm64.bin

windows-amd64:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).windows-amd64.exe

# Testing and linting targets
test:
	$(GOTEST) -v ./...

.PHONY: test-race
test-race:
	$(GOCMD) test -v --race ./...

lint:
	$(GOLINT) ./...

cover:
	$(GOCMD) test -cover ./...

clean:
	$(GOCLEAN)
	rm -f build/*.bin build/*.exe
	rm -rf $(BUILD_DIR)/Krb5Tray.app

# macOS .app bundle
APP_NAME=Krb5Tray
APP_BUNDLE=$(BUILD_DIR)/$(APP_NAME).app

.PHONY: app
app: darwin-arm64
	@echo "Creating macOS app bundle..."
	@mkdir -p $(APP_BUNDLE)/Contents/MacOS
	@mkdir -p $(APP_BUNDLE)/Contents/Resources
	@cp $(BUILD_DIR)/$(BINARY_NAME).darwin-arm64.bin $(APP_BUNDLE)/Contents/MacOS/$(APP_NAME)
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > $(APP_BUNDLE)/Contents/Info.plist
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '<plist version="1.0">' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '<dict>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>CFBundleExecutable</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <string>$(APP_NAME)</string>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>CFBundleIdentifier</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <string>com.krb5tray.app</string>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>CFBundleName</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <string>$(APP_NAME)</string>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>CFBundlePackageType</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <string>APPL</string>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>CFBundleShortVersionString</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <string>1.0.0</string>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>CFBundleVersion</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <string>$(DATE)</string>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>LSMinimumSystemVersion</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <string>11.0</string>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>LSUIElement</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <true/>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <key>NSHighResolutionCapable</key>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '    <true/>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '</dict>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo '</plist>' >> $(APP_BUNDLE)/Contents/Info.plist
	@echo "Created $(APP_BUNDLE)"
	codesign -s "SorinS_Signing" -f --deep $(APP_BUNDLE)
