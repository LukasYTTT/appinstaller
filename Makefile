BINARY_NAME=appinstall
BUILD_DIR=bin

.PHONY: all build clean

all: build

build:
	@echo "Building \$(BINARY_NAME)..."
	@go build -o \$(BINARY_NAME) .

clean:
	@echo "Cleaning up..."
	@rm -f \$(BINARY_NAME)
	@rm -rf \$(BUILD_DIR)
