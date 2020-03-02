VERSION:=$(shell git rev-list --count HEAD)-$(shell git rev-parse --short HEAD)
DATE:=$(shell date -u '+%Y-%m-%d-%H%M UTC')
BIN_DIR = $(PWD)/bin
GO := GO111MODULE=on go


.PHONY: help
help:
	@echo 'Available commands:'
	@echo '* help                - Show this message'
	@echo '* lint                - Lint code'
	@echo '* test                - Run tests'
	@echo '* clean               - Clean built binaries'
	@echo '* install			 - Install dependencies'


install:
	$(GO) mod download
	$(GO) mod vendor

.PHONY: lint
lint: FILE ?= ./...
lint:
	golangci-lint run --deadline=5m $(FILE)


.PHONY: test
test:
	go test -v -race ./...

bin/banner-api: 
	go build -o bin/banner-api ./cmd/banner-api

.PHONY: clean
clean:
	rm -rf bin vendor
