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

GIT_TAG:=$(shell git describe --tags 2> /dev/null)
GIT_TAG_STRIPPED:=$(patsubst v%,%,$(GIT_TAG))

RPM_VERSION=$(GIT_TAG_STRIPPED)
RPM_NAME=aws-service-lookup
RPM_DESC=find nodes/services from the EC2 API
RPM_URL=https://github.com/boyvinall/aws-service-lookup
RPM=$(RPM_NAME)-$(RPM_VERSION)-1.x86_64.rpm

.PHONY: rpm
rpm: $(RPM)

$(RPM): usr/bin/aws-service-lookup
	fpm -s dir -t rpm \
		-n $(RPM_NAME) \
		-v $(RPM_VERSION) \
		--description "$(RPM_DESC)" \
		--url "$(RPM_URL)" \
		etc usr

usr/bin/aws-service-lookup: aws-service-lookup
	mkdir -p $(dir $@)
	cp $< $@

clean::
	rm -rf $(RPM) usr aws-service-lookup
