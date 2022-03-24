# SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

export CGO_ENABLED=1
export GO111MODULE=on

.PHONY: build

ONOS_RSM_VERSION := latest
ONOS_BUILD_VERSION := v0.6.6
ONOS_PROTOC_VERSION := v0.6.6
BUF_VERSION := 0.27.1

build-tools:=$(shell if [ ! -d "./build/build-tools" ]; then cd build && git clone https://github.com/onosproject/build-tools.git; fi)
include ./build/build-tools/make/onf-common.mk

build: # @HELP build the Go binaries and run all validations (default)
build:
	GOPRIVATE="github.com/onosproject/*" go build -o build/_output/onos-rsm ./cmd/onos-rsm

test: # @HELP run the unit tests and source code validation
test: build deps linters license
	go test -race github.com/onosproject/onos-rsm/pkg/...
	go test -race github.com/onosproject/onos-rsm/cmd/...

jenkins-test:  # @HELP run the unit tests and source code validation producing a junit style report for Jenkins
jenkins-test: deps license linters
	TEST_PACKAGES=github.com/onosproject/onos-rsm/... ./build/build-tools/build/jenkins/make-unit

buflint: #@HELP run the "buf check lint" command on the proto files in 'api'
	docker run -it -v `pwd`:/go/src/github.com/onosproject/onos-rsm \
		-w /go/src/github.com/onosproject/onos-rsm/api \
		bufbuild/buf:${BUF_VERSION} check lint

protos: # @HELP compile the protobuf files (using protoc-go Docker)
protos:
	docker run -it -v `pwd`:/go/src/github.com/onosproject/onos-rsm \
		-w /go/src/github.com/onosproject/onos-rsm \
		--entrypoint build/bin/compile-protos.sh \
		onosproject/protoc-go:${ONOS_PROTOC_VERSION}

onos-rsm-docker: # @HELP build onos-rsm Docker image
onos-rsm-docker:
	@go mod vendor
	docker build . -f build/onos-rsm/Dockerfile \
		-t onosproject/onos-rsm:${ONOS_RSM_VERSION}
	@rm -rf vendor

images: # @HELP build all Docker images
images: build onos-rsm-docker

kind: # @HELP build Docker images and add them to the currently configured kind cluster
kind: images
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image onosproject/onos-rsm:${ONOS_RSM_VERSION}

all: build images

publish: # @HELP publish version on github and dockerhub
	./build/build-tools/publish-version ${VERSION} onosproject/onos-rsm

jenkins-publish:  # @HELP Jenkins calls this to publish artifacts
	./build/bin/push-images
	./build/build-tools/release-merge-commit

clean:: # @HELP remove all the build artifacts
	rm -rf ./build/_output ./vendor ./cmd/onos-rsm/onos-rsm ./cmd/onos/onos
	go clean -testcache github.com/onosproject/onos-rsm/...

