.PHONY: fast
fast: build coverage-short lint-fast

.PHONY: all
all: build coverage lint-full

export GOPATH:=$(realpath $(shell pwd)/../../../..)

GOMAKE:=gopkg.in/make.v3
-include $(GOPATH)/src/$(GOMAKE)/batteries.mk
$(GOPATH)/src/gopkg.in/$(GOMAKE)/batteries.mk:
	go get gopkg.in/$(GOMAKE)

.PHONY: build
build: aws-service-lookup

install: vendor

.PHONY: aws-service-lookup
aws-service-lookup: vendor
	$(call PROMPT,Building $@)
	CGO_ENABLED=0 $(GO) build -o $@ -ldflags="-s -w"
