BINARY := multiband
MODULE := codeberg.org/splitringresonator/multiband
CONTAINER_BUILD_TOOL ?= buildah

COMMIT = $(shell git rev-parse HEAD)
SHORT = $(shell git describe --tags --always --dirty=\*)
LDFLAGS := -ldflags="-X=${MODULE}/internal/version.Short=${SHORT} -X=${MODULE}/internal/version.Commit=${COMMIT} -X=${MODULE}/internal/version.RawBuildTimestamp=$(shell date +%s)"
GO_BUILD_ARGS := ${LDFLAGS} -buildvcs=true


all: bin

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: bin
bin: docs
	go build ${GO_BUILD_ARGS} -o bin/${BINARY} ${MODULE}
	# ./bin/${BINARY} -h

.PHONY: docs
docs:
	go build ${GO_BUILD_ARGS} -o bin/${BINARY} ${MODULE}
	go run ./internal/tools/docgen -out ./docs/cli -format markdown # -frontmatter
	#go run ./internal/tools/docgen -out ./content/reference -format markdown -frontmatter
	go run ./internal/tools/docgen -out ./man -format man
	#go run ./internal/tools/docgen -out ./docs/rest -format rest

.PHONY: container
container:
	$(CONTAINER_BUILD_TOOL) build -f Containerfile -t $(BINARY):latest

.PHONY: clean
clean:
	rm bin/*
