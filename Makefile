BINARY := kunai

# Determine GOBIN (where go install would put binaries)
GOBIN := $(shell go env GOBIN)
# If GOBIN is empty, default to GOPATH/bin
ifeq ($(GOBIN),)
	GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: all build install clean

all: build

# Build the project into a local binary
build:
	@echo "==> Ensuring dependencies are up-to-date"
	go mod tidy
	@echo "==> Building $(BINARY)"
	go build -o $(BINARY) .

# Move the binary into GOBIN
install: build
	@echo "==> Installing $(BINARY) to $(GOBIN)"
	mkdir -p $(GOBIN)
	mv $(BINARY) $(GOBIN)/$(BINARY)
	@echo "==> $(BINARY) installed successfully"

# Remove the built binary
clean:
	@echo "==> Cleaning up"
	rm -f $(BINARY)