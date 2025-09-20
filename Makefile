BINARY := multiband
MODULE := github.com/byxorna/multiband
BUILDER := podman

COMMIT = $(shell git rev-parse HEAD)
BUILD = $(shell git describe --tags --always --dirty=\*)
LDFLAGS := -ldflags="-X=${MODULE}/pkg/version.Build=${BUILD} -X=${MODULE}/pkg/version.Commit=${COMMIT}"
GO_BUILD_ARGS := ${LDFLAGS} -buildvcs=true


all: bin

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: bin
bin:
	go build ${GO_BUILD_ARGS} -o bin/${BINARY} ${MODULE}

.PHONY: container
container:
	$(BUILDER) build -f Containerfile -t $(BINARY):latest

.PHONY: clean
clean:
	rm bin/*
