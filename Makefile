export PATH := ${GOPATH}/bin:${PATH}
BUILD_ARGS := -a -installsuffix cgo -ldflags="-w -s"
BINARY_NAME := protodep

.PHONY: all dep compile test

all: dep compile

dep:
	go mod download

test:
	go test -v ./...

vet:
	go vet ./...

compile:
	go build $(BUILD_ARGS) -o $(BINARY_NAME)

goreleaser:
	curl -sL https://git.io/goreleaser | bash -s -- --snapshot --skip-publish --rm-dist