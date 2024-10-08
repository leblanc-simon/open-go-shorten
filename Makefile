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

prod: $(OUTDIR) ## Build a production version
	CGO_ENABLED=0 go build -ldflags "-s -w -buildid= -X 'main.version=$(version)' -trimpath -o $(OUTDIR)/open-go-shorten main.go

release: prod ## Generate a production release
	tar -cjf $(OUTDIR)/../open-go-shorten-$(VERSION).tar.bz2 -C $(OUTDIR) .
	rm -fr $(OUTDIR) || true

.PHONY: help
.DEFAULT_GOAL := help
.SHELLFLAGS += -e