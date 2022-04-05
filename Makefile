SHELL := /usr/bin/env bash

BIN_FILE := ./speedtest-exporter
GO_SOURCES := $(find $(CURDIR) -type f -name "*.go" -print)
GO_SOURCES += go.mod go.sum
GOBUILDFLAGS=CGO_ENABLED=0 GOARCH=$(GOARCH) -gcflags="all=-trimpath=$(GOPATH)" -asmflags="all=-trimpath=$(GOPATH)"
BUILD_STATIC= "true"
# Container image registry settings
IMG := lisa/speedtest-exporter
REGISTRY ?= quay.io
TAG_LATEST ?= "true"

REVISION ?= 1
VERSION ?= $(shell date +%y.%m.$(REVISION))
GITHASH=$(shell git rev-list -1 HEAD)

ARCHES = amd64 arm64

include functions.mk


default: all

all: build

.PHONY: build
build: $(BIN_FILE)
# Builds for this arch
$(BIN_FILE): test vet
	go build -ldflags "-X github.com/lisa/speedtest-go-exporter/pkg/version.GitHash=$(GITHASH) -X github.com/lisa/speedtest-go-exporter/pkg/version.Version=$(VERSION)" -o $(@) .

.PHONY: docker-build
docker-build:
	@echo "Building githash $(GITHASH) into version $(VERSION)" ;\
	for build_arch in $(ARCHES); do \
		docker build --platform=linux/$$build_arch --build-arg=GITHASH=$(GITHASH) --build-arg=VERSION=$(VERSION) --build-arg=GOARCH=$$build_arch -t $(REGISTRY)/$(IMG):$$build_arch-$(VERSION) . ;\
		$(call set_image_arch,$(REGISTRY)/$(IMG):$$build_arch-$(VERSION),$$build_arch) ;\
		if [[ $(TAG_LATEST) == "true" ]]; then \
			docker tag $(REGISTRY)/$(IMG):$$build_arch-$(VERSION) $(REGISTRY)/$(IMG):$$build_arch-latest ;\
		fi ;\
	done

.PHONY: docker-multiarch
docker-multiarch: docker-build
	@appends= ;\
	latestappends= ;\
	echo "Pushing individual images to registry" ;\
	for build_arch in $(ARCHES); do \
		appends="$${appends} $(REGISTRY)/$(IMG):$$build_arch-$(VERSION) " ;\
		latestappends="$${latestappends} $(REGISTRY)/$(IMG):$$build_arch-latest " ;\
		docker push $(REGISTRY)/$(IMG):$$build_arch-$(VERSION) ;\
	done ;\
	docker manifest create $(REGISTRY)/$(IMG):$(VERSION) $$appends ;\
	if [[ $(TAG_LATEST) == "true" ]]; then \
		docker manifest create $(REGISTRY)/$(IMG):latest $$latestappends ;\
	fi ;\
	for build_arch in $(ARCHES); do \
		docker manifest annotate $(REGISTRY)/$(IMG):$(VERSION) $(REGISTRY)/$(IMG):$$build_arch-$(VERSION) --os linux --arch $$build_arch ;\
		if [[ $(TAG_LATEST) == "true" ]]; then \
			docker manifest annotate $(REGISTRY)/$(IMG):latest $(REGISTRY)/$(IMG):$$build_arch-$(VERSION) --os linux --arch $$build_arch ;\
		fi ;\
	done ;\
	echo "Done"

docker-push: docker-build docker-multiarch
	@docker manifest push $(REGISTRY)/$(IMG):$(VERSION) ;\
	if [[ $(TAG_LATEST) == "true" ]]; then \
		docker manifest push $(REGISTRY)/$(IMG):latest ;\
	fi

.PHONY: docker-clean
docker-clean:
	for build_arch in $(ARCHES); do \
		docker rmi $(REGISTRY)/$(IMG):$$build_arch-$(VERSION) || true ;\
		docker rmi $(REGISTRY)/$(IMG):$$build_arch-latest || true ;\
	done ;\
	docker rmi $(REGISTRY)/$(IMG):latest || true ;\
	docker rmi $(REGISTRY)/$(IMG):$(VERSION) || true ;\
	rm -vrf ~/.docker/manifests/$(shell echo $(REGISTRY)/$(IMG) | tr '/' '_' | tr ':' '-')-$(VERSION)  || true ;\
	rm -vrf ~/.docker/manifests/$(shell echo $(REGISTRY)/$(IMG) | tr '/' '_' | tr ':' '-')-latest || true ;\


###################
TESTOPTS ?=

.PHONY: test
test: vet $(GO_SOURCES)
	@go test $(TESTOPTS) $(shell go list -mod=readonly -e ./...)

.PHONY: vet
vet:
	@if [[ $$(gofmt -s -l $$(go list -f '{{ .Dir }}' ./...) | wc -l | awk '{print $$1}') -gt 0 ]]; then \
		gofmt -s -d $$(go list -f '{{ .Dir }}' ./...) ;\
		exit 1 ;\
	fi ;\
	go vet ./cmd/... ./pkg/...
