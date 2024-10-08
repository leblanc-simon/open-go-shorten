VERSION := $(shell git tag | tail -n1)
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CURRENT_DIR := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
OUTDIR := ${CURRENT_DIR}/releases/${VERSION}

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

$(OUTDIR):
	rm -fr $(OUTDIR) || true
	mkdir -p $(OUTDIR)

debug: ## Build a debug version
	go build -o open-go-shorten *.go

clean-build: $(OUTDIR) ## Clean the build directory

build-linux: ## Build release version for GNU/Linux
	@GOOS="linux" GOARCH="amd64" CGO_ENABLED=0 go build -ldflags "-s -w -buildid= -X 'main.version=$(VERSION)'" -trimpath -o $(OUTDIR)/open-go-shorten-linux-amd64 main.go
	@GOOS="linux" GOARCH="arm64" CGO_ENABLED=0 go build -ldflags "-s -w -buildid= -X 'main.version=$(VERSION)'" -trimpath -o $(OUTDIR)/open-go-shorten-linux-arm64 main.go

build-darwin: ## Build release version for MacOS
	@GOOS="darwin" GOARCH="amd64" CGO_ENABLED=0 go build -ldflags "-s -w -buildid= -X 'main.version=$(VERSION)'" -trimpath -o $(OUTDIR)/open-go-shorten-darwin-amd64 main.go
	@GOOS="darwin" GOARCH="arm64" CGO_ENABLED=0 go build -ldflags "-s -w -buildid= -X 'main.version=$(VERSION)'" -trimpath -o $(OUTDIR)/open-go-shorten-darwin-arm64 main.go

build-windows: ## Build release version for MacOS
	@GOOS="windows" GOARCH="amd64" CGO_ENABLED=0 go build -ldflags "-s -w -buildid= -X 'main.version=$(VERSION)'" -trimpath -o $(OUTDIR)/open-go-shorten-windows-amd64.exe main.go

release: clean-build build-linux build-darwin build-windows ## Build the release version

.PHONY: help
.DEFAULT_GOAL := help
.SHELLFLAGS += -e