# Variables
GOHOSTOS:=$(shell go env GOHOSTOS)
GOHOSTARCH:=$(shell go env GOHOSTARCH)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse HEAD)

# Docker variables
DOCKER_REGISTRY?=sn0wl1n
IMAGE_NAME?=ech0
IMAGE_TAG?=latest
OS?=$(if $(GOHOSTOS),$(GOHOSTOS),linux)
ARCH?=$(if $(GOHOSTARCH),$(GOHOSTARCH),amd64)

.PHONY: help air-install run dev web-dev lint fmt test wire build-image push-image

help:
	@echo "Available targets:"
	@echo "  make run         - Run backend in web mode"
	@echo "  make dev         - Run backend with Air hot reload"
	@echo "  make air-install - Install Air to GOPATH/bin"
	@echo "  make web-dev     - Run frontend dev server"
	@echo "  make lint        - Run golangci-lint checks"
	@echo "  make fmt         - Run golangci-lint formatters"
	@echo "  make test        - Run Go tests"
	@echo "  make wire        - Generate DI code via Wire"
	@echo "  make build-image - Build Docker image"
	@echo "  make push-image  - Push Docker image"

air-install:
	go install github.com/air-verse/air@latest

run:
	go run ./main.go web

dev:
	air -c .air.toml

web-dev:
	cd web && pnpm dev

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

test:
	go test ./...

wire:
	cd internal/di && wire
build-image:
	@echo "Building image for platform: $(OS)/$(ARCH)"
	docker build --platform $(OS)/$(ARCH) \
		--build-arg TARGETOS=$(OS) \
		--build-arg TARGETARCH=$(ARCH) \
		-t $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) -f build.Dockerfile .
push-image:
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

