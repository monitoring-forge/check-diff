VERSION=0.0.7
GITCOMMIT?=$(shell git describe --dirty --always)
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.commit=${GITCOMMIT}"

all: check-diff

.PHONY: check-diff

check-diff: main.go open_unix.go open_windows.go
	go build $(LDFLAGS) -o check-diff .

linux: main.go open_unix.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o check-diff open_unix.go main.go

check:
	go test -v ./...

