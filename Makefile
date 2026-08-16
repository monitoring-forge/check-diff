VERSION=0.0.7
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"

all: check-diff

.PHONY: check-diff lint check linux

check-diff: *.go
	go build $(LDFLAGS) -o check-diff

linux: *.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o check-diff

check:
	go test -v ./...

lint:
	golangci-lint run ./...