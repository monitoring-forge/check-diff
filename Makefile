VERSION=0.0.5
GITCOMMIT?=$(shell git describe --dirty --always)
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.commit=${GITCOMMIT}"

all: check-diff

.PHONY: check-diff

check-diff: main.go
	go build $(LDFLAGS) -o check-diff main.go

linux: main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o check-diff main.go

check:
	go test -v ./...

