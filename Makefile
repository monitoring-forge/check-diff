VERSION=0.0.6
GITCOMMIT?=$(shell git describe --dirty --always)
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.commit=${GITCOMMIT}"

all: check-diff

.PHONY: check-diff

check-diff: *.go
	go build $(LDFLAGS) -o check-diff

linux: *.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o check-diff

check:
	go test -v ./...

lint:
	golangci-lint run ./...