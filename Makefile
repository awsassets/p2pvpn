PREFIX := p2pvpn
MODULE := github.com/lp2p/p2pvpn

BUILD_DIR   := build
BUILD_FLAGS := -v

CGO_ENABLED := 0
GO111MODULE := on

LDFLAGS += -w -s -buildid=

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

GO_BUILD = GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) \
	go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -trimpath

.PHONY: client server clean

all: client server

client:
	$(GO_BUILD) -o $(BUILD_DIR)/$(PREFIX)-$@-$(GOOS)-$(GOARCH) cmd/$@/main.go

server:
	$(GO_BUILD) -o $(BUILD_DIR)/$(PREFIX)-$@-$(GOOS)-$(GOARCH) cmd/$@/main.go

clean:
	rm -rf $(BUILD_DIR)
